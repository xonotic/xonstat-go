package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gitlab.com/xonotic/xonstat/pkg/skill"
)

// updateSkillsCmd updates the skills in the database (or simulates the same).
var updateSkillsCmd = &cobra.Command{
	Use:   "updateskills",
	Short: "Update player skills",
	Long:  `Update player skill values based on their performance in their latest games`,
	Run: func(cmd *cobra.Command, args []string) {
		start, _ := cmd.Flags().GetInt("start")
		end, _ := cmd.Flags().GetInt("end")
		limit, _ := cmd.Flags().GetInt("limit")
		resume, _ := cmd.Flags().GetBool("resume")
		resumeFile, _ := cmd.Flags().GetString("resumefile")
		simulate, _ := cmd.Flags().GetBool("simulate")

		if start == models.BlankStartingGameID && !resume {
			log.Fatal("You must provide a starting game ID if you're not resuming")
			return
		}
		updateSkills(start, end, limit, resume, resumeFile, simulate)
	},
}

func init() {
	// set up logging
	err := initLog()
	if err != nil {
		log.Fatal("Unable to initialize logging.")
	}

	rootCmd.AddCommand(updateSkillsCmd)
	updateSkillsCmd.Flags().IntP("start", "s", models.BlankStartingGameID, "Starting game ID")
	updateSkillsCmd.Flags().IntP("end", "e", models.BlankEndingGameID, "Ending game ID")
	updateSkillsCmd.Flags().IntP("limit", "l", models.BlankLimit, "Limit the number of games processed")
	updateSkillsCmd.Flags().BoolP("resume", "r", true, "Resume from where we last left off (use resumefile contents)")
	updateSkillsCmd.Flags().String("resumefile", "skill_state.txt", "File containing the game ID to start from")
	updateSkillsCmd.Flags().Bool("simulate", false, "Do not change the database")
}

func updateSkills(start, end, limit int, resume bool, resumeFile string, simulate bool) {
	dsn := viper.GetString("ConnStr")
	db, err := models.NewPGDatastore(dsn)
	if err != nil {
		log.Fatal("Unable to initialize database connection.")
	}

	begin := time.Now()
	games, err := db.RGamesByRange(start, end, limit)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Collected info for %d games in %s.\n", len(games), time.Since(begin))

	begin = time.Now()

	// skills indexed by player_id-game_type_cd
	skillsByPlayer := make(map[string]*skill.Rating)

	for _, game := range games {
		rawResults, err := db.RMatchResultsByGameID(game.GameID)
		if err != nil {
			log.Printf("Error processing game %d: %s", game.GameID, err)
		}

		matchResult := skill.MatchResult{
			MatchID:       game.GameID,
			PlayerResults: make([]skill.PlayerResult, 0, len(rawResults)),
		}

		skills := make([]skill.Rating, len(rawResults))

		// For each entry in the raw results, we transform it into a format that the
		// skill package expects.
		for i, result := range rawResults {
			// First we add the player result values.
			playerResult := skill.PlayerResult{
				PlayerID: result.PlayerID,
				Score:    float32(result.Score),
			}
			matchResult.PlayerResults = append(matchResult.PlayerResults, playerResult)

			// Then we add the skill/rating values at the matching indices.
			var playerRating skill.Rating

			key := fmt.Sprintf("%d-%s", result.PlayerID, result.GameTypeCd)

			if rating, ok := skillsByPlayer[key]; ok {
				// If we've seen this player before in our calculations, we need to use that
				// Rating value since it has probably been updated. No need to update the skill map.
				playerRating = *rating
			} else {
				// Otherwise we have not seen this player before and will need to save it in the skill map
				// when done.
				if result.Mu.Valid && result.Sigma.Valid {
					// We haven't seen this player before and they have a non-null rating in the DB. Use it.
					playerRating.Mu = result.Mu.Float64
					playerRating.Sigma = result.Sigma.Float64
				} else {
					// We haven't seen this player before and the do not have a rating in the DB.
					playerRating.Mu = skill.MU
					playerRating.Sigma = skill.SIGMA

					// TODO: log that we need to insert this record when done. Maybe a set?
				}
				skillsByPlayer[key] = &playerRating
			}

			skills[i] = playerRating
		}

		// Now that we've assembled the input, let's calculate the skill updates!
		newSkills, err := skill.WengLinBT(matchResult, skills)
		if err != nil {
			log.Printf("Problem calculating Weng-Lin for game %d", game.GameID)
		}

		// Update our skill map for the next go-round. The new skills above are index-aligned
		// with the PlayerResults list that we provided in the MatchResult.
		for i, newSkill := range newSkills {
			key := fmt.Sprintf("%d-%s", matchResult.PlayerResults[i].PlayerID, game.GameTypeCd)
			skillsByPlayer[key].Mu = newSkill.Mu
			skillsByPlayer[key].Sigma = newSkill.Sigma
		}
	}
	fmt.Printf("Processed %d games in %s.\n", len(games), time.Since(begin))

	fmt.Printf("New Skills:\n")
	for key, value := range skillsByPlayer {
		fmt.Printf("%s %+v\n", key, value)
	}
}
