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

// prepareInput takes the raw data from the database and transforms it into the format that the skill
// package requires. It also takes the "running" skills maps so we can provide the algorithm with the
// most recently calculated ratings.
func prepareInput(rawResults []*models.PlayerSkillMatchResult,
	skillsByPlayer map[string]*models.PlayerSkill, brandNewKeys map[string]struct{}) (skill.MatchResult, []skill.Rating) {

	matchResult := skill.MatchResult{
		MatchID:       rawResults[0].GameID,
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
			playerRating.Mu = rating.Mu
			playerRating.Sigma = rating.Sigma
		} else {
			// Otherwise we have not seen this player before and will need to save it in the skill map
			// when done.
			if result.Mu.Valid && result.Sigma.Valid {
				// We haven't seen this player before but they have a rating in the DB. Use it.
				playerRating.Mu = result.Mu.Float64
				playerRating.Sigma = result.Sigma.Float64
			} else {
				// We haven't seen this player before and the do not have a rating in the DB.
				playerRating.Mu = skill.MU
				playerRating.Sigma = skill.SIGMA

				brandNewKeys[key] = struct{}{}
			}
			skillsByPlayer[key] = &models.PlayerSkill{
				PlayerID:   result.PlayerID,
				GameTypeCd: result.GameTypeCd,
				Mu:         playerRating.Mu,
				Sigma:      playerRating.Sigma,
				ActiveInd:  true,
			}
		}

		skills[i] = playerRating
	}

	return matchResult, skills
}

// updateSkills calculates new skill ratings for a range of games.
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

	log.Printf("Collected info for %d games in %s.\n", len(games), time.Since(begin))

	begin = time.Now()

	// skills indexed by player_id-game_type_cd
	// TODO: This needs to be updated with a PlayerSkill object when ready.
	skillsByPlayer := make(map[string]*models.PlayerSkill)

	// Used to keep track of which keys (playerID + gameTypeCd) are brand new, thus
	// need to be inserted rather than updated at the end of processing.
	brandNewKeys := make(map[string]struct{})

	for _, game := range games {
		rawResults, err := db.RMatchResultsByGameID(game.GameID)
		if err != nil {
			log.Printf("Error processing game %d: %s", game.GameID, err)
		}

		// Don't even bother for games with only one legit player.
		if len(rawResults) < 2 {
			continue
		}

		// Assemble the input the way skill.WengLinBT expects.
		matchResult, skills := prepareInput(rawResults, skillsByPlayer, brandNewKeys)

		// Calculate the skill updates, returning a new skill list.
		newSkills, err := skill.WengLinBT(matchResult, skills)
		if err != nil {
			log.Printf("Problem calculating Weng-Lin for game %d", game.GameID)
		}

		// Update our skill map for the next go-round. The new skills above are index-aligned
		// with the PlayerResults list that we provided in the MatchResult.
		for i, newSkill := range newSkills {
			key := fmt.Sprintf("%d-%s", matchResult.PlayerResults[i].PlayerID, game.GameTypeCd)

			// TODO: Determine minimum values here. Prevent gonig below that.
			skillsByPlayer[key].Mu = newSkill.Mu
			skillsByPlayer[key].Sigma = newSkill.Sigma
		}
	}
	log.Printf("Processed %d games in %s.\n", len(games), time.Since(begin))

	if simulate {
		log.Printf("Need to create %d new records.", len(brandNewKeys))
		log.Printf("New Skills:")
		for key, value := range skillsByPlayer {
			log.Printf("%s %+v\n", key, value)
		}
	}
}
