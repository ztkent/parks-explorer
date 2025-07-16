package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
	"github.com/ztkent/nps-dashboard/internal/dashboard"
	"github.com/ztkent/replay"
)

func main() {
	// Set default database path if not provided
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/dashboard.db"
	}

	dashManager := dashboard.NewDashboard(os.Getenv("NPS_API_KEY"), dbPath)
	// Initialize router and middleware
	r := chi.NewRouter()
	// Log request and recover from panics
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Define routes
	DefineRoutes(r, dashManager, replay.NewCache(
		replay.WithMaxSize(100),
		replay.WithMaxMemory(1024*1024*1024),
		replay.WithEvictionPolicy("LRU"),
		replay.WithTTL(5*time.Minute),
		replay.WithMaxTTL(30*time.Minute),
		replay.WithCacheFilters([]string{"URL", "Method"}),
		replay.WithLogger(log.New(os.Stdout, "replay: ", log.LstdFlags)),
	))

	// Start server
	fmt.Println("Dashboard is running on port " + os.Getenv("SERVER_PORT"))
	if os.Getenv("ENV") == "dev" {
		log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), cors.Default().Handler(r)))
	}
	log.Fatal(http.ListenAndServeTLS(":"+os.Getenv("SERVER_PORT"), os.Getenv("CERT_PATH"), os.Getenv("CERT_KEY_PATH"), r))
}

func DefineRoutes(r *chi.Mux, dashManager *dashboard.Dashboard, cache *replay.Cache) {
	// Apply a rate limiter to all routes
	r.Use(httprate.Limit(
		50,             // requests
		60*time.Second, // per durations
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))

	// Apply visitor tracking middleware
	r.Use(dashManager.TagVistorsMiddleware)

	// Static routes
	r.Get("/", dashManager.HomeHandler())
	r.Get("/parks/{slug}", cache.MiddlewareFunc(dashManager.ParkPageHandler))
	r.Get("/static/*", dashManager.StaticFileHandler())

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes
		r.Get("/auth/google", dashManager.GoogleLoginHandler)
		r.Get("/auth/callback", dashManager.GoogleCallbackHandler)
		r.Get("/auth/logout", dashManager.LogoutHandler)
		r.Get("/user-info", dashManager.UserInfoHandler)
		r.Get("/avatar", dashManager.AvatarProxyHandler)

		// HTMX routes
		r.Get("/auth-status", dashManager.AuthStatusHandler)
		r.Get("/parks", cache.MiddlewareFunc(dashManager.ParksHandler))
		r.Get("/parks/featured", cache.MiddlewareFunc(dashManager.FeaturedParksHandler))
		r.Get("/parks/search", cache.MiddlewareFunc(dashManager.ParkSearchHandler))

		// Park tab content routes
		r.Get("/parks/{parkCode}/overview", cache.MiddlewareFunc(dashManager.ParkOverviewHandler))
		r.Get("/parks/{parkCode}/activities", cache.MiddlewareFunc(dashManager.ParkActivitiesHandler))
		r.Get("/parks/{parkCode}/media", cache.MiddlewareFunc(dashManager.ParkMediaHandler))
		r.Get("/parks/{parkCode}/news", cache.MiddlewareFunc(dashManager.ParkNewsHandler))
		r.Get("/parks/{parkCode}/details", cache.MiddlewareFunc(dashManager.ParkDetailsHandler))

		// Template routes
		r.Get("/templates/{template}", dashManager.TemplateHandler)
	})
}
