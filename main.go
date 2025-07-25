package main

import (
	"crypto/tls"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "embed"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ztkent/parks-explorer/internal/dashboard"
	"github.com/ztkent/replay"
)

//go:embed certs/parks_cert.pem
var certPEM []byte

//go:embed certs/parks_key.pem
var keyPEM []byte

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
		replay.WithMaxMemory(1024*1024*1024*100),
		replay.WithEvictionPolicy("LRU"),
		replay.WithTTL(6*time.Hour),
		replay.WithMaxTTL(24*time.Hour),
		replay.WithCacheFilters([]string{"URL", "Method"}),
		replay.WithLogger(log.New(os.Stdout, "replay: ", log.LstdFlags)),
	))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8086" // Default port
	}

	fmt.Println("Starting server on port", port)
	if os.Getenv("ENV") == "dev" {
		// Development mode - serve HTTP only
		log.Fatal(http.ListenAndServe(":"+port, r))
	} else {
		// Production mode - serve HTTPS with embedded certificates
		cert, err := tls.X509KeyPair(certPEM, keyPEM)
		if err != nil {
			log.Fatal("Failed to load embedded certificates:", err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		server := &http.Server{
			Addr:      ":" + port,
			Handler:   r,
			TLSConfig: tlsConfig,
		}
		log.Fatal(server.ListenAndServeTLS("", ""))
	}
}

func DefineRoutes(r *chi.Mux, dashManager *dashboard.Dashboard, cache *replay.Cache) {
	// Apply visitor tracking middleware
	r.Use(dashManager.TagVistorsMiddleware)

	// Static routes
	r.Get("/", dashManager.HomeHandler())
	r.Get("/things-to-do", cache.MiddlewareFunc(dashManager.ThingsToDoPageHandler))
	r.Get("/events", cache.MiddlewareFunc(dashManager.EventsPageHandler))
	r.Get("/camping", cache.MiddlewareFunc(dashManager.CampingPageHandler))
	r.Get("/news", cache.MiddlewareFunc(dashManager.NewsPageHandler))
	r.Get("/parks/{slug}", cache.MiddlewareFunc(dashManager.ParkPageHandler))

	// Serve specific files from static directory at top level
	r.Get("/robots.txt", dashManager.TopLevelStaticFileHandler("robots.txt"))
	r.Get("/site.webmanifest", dashManager.TopLevelStaticFileHandler("site.webmanifest"))
	r.Get("/sitemap.xml", dashManager.TopLevelStaticFileHandler("sitemap.xml"))

	r.Get("/static/*", dashManager.StaticFileHandler())

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Auth routes
		r.Get("/auth/google", dashManager.GoogleLoginHandler)
		r.Get("/auth/google/callback", dashManager.GoogleCallbackHandler)
		r.Get("/auth/logout", dashManager.LogoutHandler)
		r.Get("/user-info", dashManager.UserInfoHandler)
		r.Get("/avatar", dashManager.AvatarProxyHandler)

		// Analytics routes
		r.Get("/analytics/config", dashManager.AnalyticsConfigHandler)

		// Image proxy route for secure image serving
		r.Get("/image-proxy", cache.MiddlewareFunc(dashManager.ImageProxyHandler))

		// HTMX routes
		r.Get("/auth-status", dashManager.AuthStatusHandler)
		r.Get("/parks", cache.MiddlewareFunc(dashManager.ParksHandler))
		r.Get("/parks/featured", cache.MiddlewareFunc(dashManager.FeaturedParksHandler))
		r.Get("/parks/search", cache.MiddlewareFunc(dashManager.ParkSearchHandler))

		// Things To Do routes
		r.Get("/things-to-do/search", cache.MiddlewareFunc(dashManager.ThingsToDoSearchHandler))

		// Events routes
		r.Get("/events/search", cache.MiddlewareFunc(dashManager.EventsSearchHandler))
		r.Get("/events/{eventID}/details", cache.MiddlewareFunc(dashManager.EventDetailsHandler))

		// Camping routes
		r.Get("/camping/search", cache.MiddlewareFunc(dashManager.CampingSearchHandler))

		// News routes
		r.Get("/news/search", cache.MiddlewareFunc(dashManager.NewsSearchHandler))

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
