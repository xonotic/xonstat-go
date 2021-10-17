package cmd

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ungerik/go-cairo"
	"gitlab.com/xonotic/xonstat/pkg/badges"
)

// pngToJpg saves space by converting the given PNG into a JPG with the provided quality.
func pngToJpg(pngFilename string, quality int) {
	jpgFilename := strings.Replace(pngFilename, ".png", ".jpg", -1)

	f, err := os.Open(pngFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	image, err := png.Decode(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	outfile, err := os.Create(jpgFilename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outfile.Close()

	opts := jpeg.Options{Quality: quality}

	err = jpeg.Encode(outfile, image, &opts)
	if err != nil {
		fmt.Println(err)
		return
	}

	os.Remove(pngFilename)
}

func renderWorker(pids <-chan int, wg *sync.WaitGroup, pp *badges.PlayerDataFetcher, skins map[string]badges.Skin,
	surfaceCache map[string]*cairo.Surface, outputDir string) {

	for pid := range pids {
		pd, err := pp.GetPlayerData(pid)
		if err != nil {
			log.Println(err)
		}

		if len(pd.Nick) == 0 {
			log.Printf("No data for player #%d!\n", pid)
		} else {
			for name, skin := range skins {
				pngFN := fmt.Sprintf("%s/%s/%d.png", outputDir, name, pid)
				if name == "default" {
					// The default skin doesn't go into a subdirectory.
					pngFN = fmt.Sprintf("%s/%d.png", outputDir, pid)
				}
				skin.Render(pd, pngFN, surfaceCache)
				pngToJpg(pngFN, 90)
			}
		}
	}

	wg.Done()
}

// badgesCmd generates images from player stats called "badges"
var badgesCmd = &cobra.Command{
	Use:   "badges",
	Short: "Generate badge images from player stats",
	Run: func(cmd *cobra.Command, args []string) {
		flags := cmd.Flags()
		all, _ := flags.GetBool("all")
		delta, _ := flags.GetInt("delta")
		pid, _ := flags.GetInt("pid")
		limit, _ := flags.GetInt("limit")
		workers, _ := flags.GetInt("workers")
		outputDir, _ := flags.GetString("out")

		dsn := viper.GetString("ConnStr")
		pp, err := badges.NewPlayerDataFetcher(dsn)
		if err != nil {
			log.Fatal(err)
		}

		// negate the delta value if we want all players generated
		if all {
			delta = -1
		}

		var pids []int
		if pid == -1 {
			// Find players by delta, with an optional limit
			pids, err = pp.FindPlayers(delta, limit)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// Use just the one player ID for fetching
			log.Printf("Generating a badge for just player #%d.\n", pid)
			pids = []int{pid}
		}

		cwd, _ := os.Getwd()
		skinDir := path.Join(cwd, "/assets/skins")

		log.Printf("Loading skins from '%s'", skinDir)
		skins := badges.LoadSkins(skinDir)
		for name := range skins {
			err := os.MkdirAll(fmt.Sprintf("%s/%s", outputDir, name), os.FileMode(0755))
			if err != nil {
				fmt.Println(err)
			}
		}
		log.Printf("Loaded %d skins.", len(skins))

		surfaceCache := badges.LoadSurfaces(skins)

		pidsChan := make(chan int)

		var wg sync.WaitGroup

		// start workers
		for w := 1; w <= workers; w++ {
			wg.Add(1)
			go renderWorker(pidsChan, &wg, pp, skins, surfaceCache, outputDir)
		}

		// send them work
		for _, pid := range pids {
			pidsChan <- pid
		}
		close(pidsChan)

		// wait until they are all done
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(badgesCmd)
	badgesCmd.Flags().BoolP("all", "a", false, "Generate for all players")
	badgesCmd.Flags().IntP("delta", "d", 6, "Generate for players having activity within this number of hours")
	badgesCmd.Flags().IntP("pid", "p", -1, "Generate just for this player ID")
	badgesCmd.Flags().IntP("limit", "l", -1, "Max number of badges to generate")
	badgesCmd.Flags().IntP("workers", "w", 5, "Number of worker threads")
	badgesCmd.Flags().StringP("out", "o", "output", "Output directory")
}
