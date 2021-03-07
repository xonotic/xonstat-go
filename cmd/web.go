package cmd

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/alehano/reverse"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/xonotic/xonstat/internal/handlers"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/swaggo/http-swagger"
)

// @title XonStat API
// @version 1.0
// @description JSON and textual API for Xonotic statistics.
// @termsOfService http://xonotic.org/tos

// @contact.name Support
// @contact.url https://gitlab.com/xonotic/xonstat-go/-/issues
// @contact.email antibody@xonotic.org

// @license.name AGPL 3.0
// @license.url https://www.gnu.org/licenses/agpl-3.0.en.html

// @host https://stats.xonotic.org
// @BasePath /
// @query.collection.format multi

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

func web(addr string) {
	dsn := viper.GetString("ConnStr")
	db, err := models.NewPGDatastore(dsn)
	if err != nil {
		log.Fatal("Unable to initialize database connection.")
	}

	requestLogger := lumberjack.Logger{
		Filename:   viper.GetString("RequestsLogFile"),
		MaxSize:    viper.GetInt("RequestsMaxSize"),
		MaxBackups: viper.GetInt("RequestsMaxBackups"),
		MaxAge:     viper.GetInt("RequestsMaxAge"),
		Compress:   true,
	}

	env := handlers.NewAppEnv(db, &requestLogger)

	r := chi.NewRouter()

	// Save the real IP address from X-Forward-For and the like.
	r.Use(middleware.RealIP)

	// Recover from panics
	r.Use(middleware.Recoverer)

	// Log request metadata: the URI, the response, and how long it took.
	formatter := middleware.DefaultLogFormatter{Logger: logger, NoColor: true}
	middleware.DefaultLogger = middleware.RequestLogger(&formatter)
	r.Use(middleware.Logger)

	// Support for low-impact uptime testing from external services
	heartbeat := middleware.Heartbeat("/ping")
	r.Use(heartbeat)

	// Routing group that verifies all requests with d0_blind_id (if enabled).
	r.Group(func(r chi.Router) {
		if viper.GetBool("VerifyRequests") {
			r.Use(handlers.D0Verify)
		}

		r.Post("/stats/submit", env.SubmissionHandler)
		r.Post("/player/me", env.PlayerHashkeyInfoHandler)
		r.Get("/elo", env.PlayerEloInfoHandler)
	})

	// Register all "regular" routes and handlers.
	r.Get(reverse.Add("leaderboard", "/"), env.LeaderboardHandler)
	r.Get(reverse.Add("summary", "/summary"), env.SummaryHandler)
	r.Get(reverse.Add("topactive", "/topactive"), env.TopActiveHandler)
	r.Get(reverse.Add("topservers", "/topservers"), env.TopServersHandler)
	r.Get(reverse.Add("topmaps", "/topmaps"), env.TopMapsHandler)
	r.Get(reverse.Add("games", "/games"), env.RecentGamesHandler)
	r.Get(reverse.Add("server_index", "/servers"), env.ServerIndexHandler)
	r.Get(reverse.Add("server_info", "/server/{id:\\d+}", "{id:\\d+}"), env.ServerInfoHandler)
	r.Get(reverse.Add("server_top_scorers", "/server/{id:\\d+}/topscorers", "{id:\\d+}"), env.ServerTopScorersHandler)
	r.Get(reverse.Add("map_index", "/maps"), env.MapIndexHandler)
	r.Get(reverse.Add("map_info", "/map/{id:\\d+}", "{id:\\d+}"), env.MapInfoHandler)
	r.Get(reverse.Add("game_info", "/game/{id:\\d+}", "{id:\\d+}"), env.GameInfoHandler)
	r.Get(reverse.Add("game_weapon_info", "/game/{id:\\d+}/weapons", "{id:\\d+}"), env.GameWeaponInfoHandler)
	r.Get(reverse.Add("player_index", "/players"), env.PlayerIndexHandler)
	r.Get(reverse.Add("player_info", "/player/{id:\\d+}", "{id:\\d+}"), env.PlayerInfoHandler)
	r.Get(reverse.Add("player_weapon_info", "/player/{id:\\d+}/weapons", "{id:\\d+}"), env.PlayerWeaponInfoHandler)
	r.Get(reverse.Add("player_recent_games_fragment", "/player/{id:\\d+}/recentGamesFragment", "{id:\\d+}"), env.PlayerRecentGamesFragmentHandler)
	r.Get(reverse.Add("player_skill", "/player/{id:\\d+}/skill", "{id:\\d+}"), env.PlayerSkillHandler)
	r.Get("/skill", env.PlayerSkillHashkeyHandler)
	r.NotFound(env.NotFoundHandler)

	// Static files
	cwd, _ := os.Getwd()
	staticDir := http.Dir(filepath.Join(cwd, "web/static"))
	FileServer(r, "/static", staticDir)

	// Swagger documentation via "swag" and "swag-http" libraries.
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:6543/static/swagger.json"), //The url pointing to API definition"
	))

	// Start the web application server on the specified port.
	log.Printf("Starting XonStat web application server on %s...", addr)
	http.ListenAndServe(addr, r)
}

// webCmd starts up the web application server
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Run the web application server",
	Long:  `Run the XonStat web application server.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr, _ := cmd.Flags().GetString("addr")
		web(addr)
	},
}

func init() {
	// set up logging
	err := initLog()
	if err != nil {
		log.Fatal("Unable to initialize logging.")
	}

	rootCmd.AddCommand(webCmd)
	webCmd.Flags().StringP("addr", "a", "0.0.0.0:6543", "address")
}
