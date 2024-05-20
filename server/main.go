package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ztkent/nps-dashboard/internal/dashboard"
	"github.com/Ztkent/nps-dashboard/replay"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/rs/cors"
)

func main() {
	dashManager := dashboard.NewDashboard(os.Getenv("NPS_API_KEY"))
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
	fmt.Println("Server is running on port " + os.Getenv("SERVER_PORT"))
	if os.Getenv("ENV") == "dev" {
		log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), cors.Default().Handler(r)))
	}
	log.Fatal(http.ListenAndServeTLS(":"+os.Getenv("SERVER_PORT"), os.Getenv("CERT_PATH"), os.Getenv("CERT_KEY_PATH"), r))
	return
}

func DefineRoutes(r *chi.Mux, dashManager *dashboard.Dashboard, cache *replay.Cache) {
	// Apply a rate limiter to all routes

	r.Use(httprate.Limit(
		50,             // requests
		60*time.Second, // per durations
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))
	r.Get("/park-cams", cache.Middleware(dashManager.LiveParkCamsHandler()))
	r.Get("/park-list", cache.Middleware(dashManager.ParkListHandler()))
}
