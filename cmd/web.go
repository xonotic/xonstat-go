package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/alehano/reverse"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/httprate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/swaggo/http-swagger"
	"gitlab.com/xonotic/xonstat/internal/handlers"
	"gitlab.com/xonotic/xonstat/pkg/models"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/go-redis/redis/v8"
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

// @schemes https
// @host stats.xonotic.org
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

	redisAddr := viper.GetString("RedisAddr")
	cache, err := models.NewRedisCache(&redis.Options{
		Addr: redisAddr,
	})

	cacheEnabled := true
	if err != nil {
		cacheEnabled = false
	}

	requestLogger := lumberjack.Logger{
		Filename:   viper.GetString("RequestsLogFile"),
		MaxSize:    viper.GetInt("RequestsMaxSize"),
		MaxBackups: viper.GetInt("RequestsMaxBackups"),
		MaxAge:     viper.GetInt("RequestsMaxAge"),
		Compress:   true,
	}

	env := handlers.NewAppEnv(db, cacheEnabled, cache, &requestLogger)

	r := chi.NewRouter()

	// Save the real IP address from X-Forward-For and the like.
	r.Use(middleware.RealIP)

	// Recover from panics
	r.Use(middleware.Recoverer)

	// Rate limiting
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

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
		r.Post("/balance", env.BalanceHandler)
	})

	// Register all "regular" routes and handlers.
	r.Get(reverse.Add("leaderboard", "/"),
		env.Cached(5*time.Minute, env.LeaderboardHandler))
	r.Get(reverse.Add("summary", "/summary"),
		env.Cached(1*time.Hour, env.SummaryHandler))
	r.Get(reverse.Add("topactive", "/topactive"),
		env.Cached(1*time.Hour, env.TopActiveHandler))
	r.Get(reverse.Add("topservers", "/topservers"),
		env.Cached(1*time.Hour, env.TopServersHandler))
	r.Get(reverse.Add("topmaps", "/topmaps"),
		env.Cached(1*time.Hour, env.TopMapsHandler))
	r.Get(reverse.Add("games", "/games"),
		env.RecentGamesHandler)
	r.Get(reverse.Add("server_index", "/servers"),
		env.ServerIndexHandler)
	r.Get(reverse.Add("server_info", "/server/{id:\\d+}", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.ServerInfoHandler))
	r.Get(reverse.Add("server_top_scorers", "/server/{id:\\d+}/topscorers", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.ServerTopScorersHandler))
	r.Get(reverse.Add("map_index", "/maps"),
		env.MapIndexHandler)
	r.Get(reverse.Add("map_info", "/map/{id:\\d+}", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.MapInfoHandler))
	r.Get(reverse.Add("game_info", "/game/{id:\\d+}", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.GameInfoHandler))
	r.Get(reverse.Add("game_weapon_info", "/game/{id:\\d+}/weapons", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.GameWeaponInfoHandler))
	r.Get(reverse.Add("player_index_fragment", "/playerIndexFragment"),
		env.PlayerIndexHandler)
	r.Get(reverse.Add("player_index", "/players"),
		env.PlayerIndexHandler)
	r.Get(reverse.Add("player_info", "/player/{id:\\d+}", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.PlayerInfoHandler))
	r.Get(reverse.Add("player_weapon_info", "/player/{id:\\d+}/weapons", "{id:\\d+}"),
		env.Cached(5*time.Minute, env.PlayerWeaponInfoHandler))
	r.Get(reverse.Add("player_recent_games_fragment", "/player/{id:\\d+}/recentGamesFragment", "{id:\\d+}"),
		env.PlayerRecentGamesFragmentHandler)
	r.Get(reverse.Add("player_skill", "/player/{id:\\d+}/skill", "{id:\\d+}"),
		env.PlayerSkillHandler)
	r.Get("/skill",
		env.PlayerSkillHashkeyHandler)
	r.NotFound(env.NotFoundHandler)

	// Static files
	cwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	staticDir := http.Dir(filepath.Join(cwd, "web/static"))
	FileServer(r, "/static", staticDir)

	r.Get(reverse.Add("robots", "/robots.txt"), func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(cwd, "web/static/robots.txt"))
	})

	// Swagger documentation via "swag" and "swag-http" libraries.
	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/static/swagger.json"), //The url pointing to API definition"
	))

	// Start the web application server on the specified port.
	log.Printf("Starting XonStat web application server on %s...", addr)

	// Graceful shutdown courtesy of https://millhouse.dev/posts/graceful-shutdowns-in-golang-with-signal-notify-context.
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Construct our server and start it in the background.
	server := http.Server{
		Addr:    addr,
		Handler: r,
	}
	go server.ListenAndServe()

	// Wait for the signal to stop, then do so gracefully. The maximum shutdown time
	// is 30 seconds, after which all pending requests are cancelled.
	<-ctx.Done()
	stop()
	log.Printf("Received the interrupt/terminate signal. Shutting down gracefully. Control+C again to force.")

	// Perform application shutdown with a maximum timeout of 30 seconds.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}
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
