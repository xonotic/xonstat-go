package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/pkg/models"
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

	elapsed := time.Since(begin)
	fmt.Printf("Collected info for %d games in %s.\n", len(games), elapsed)
}
