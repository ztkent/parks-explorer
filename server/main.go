package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Ztkent/nps-dashboard/internal/dashboard"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

func main() {
	dashManager := dashboard.NewDashboard(os.Getenv("NPS_API_KEY"))
	// Initialize router and middleware
	r := chi.NewRouter()
	// Log request and recover from panics
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Define routes
	DefineRoutes(r, dashManager)

	// Start server
	fmt.Println("Server is running on port " + os.Getenv("SERVER_PORT"))
	if os.Getenv("ENV") == "dev" {
		log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVER_PORT"), r))
	}
	log.Fatal(http.ListenAndServeTLS(":"+os.Getenv("SERVER_PORT"), os.Getenv("CERT_PATH"), os.Getenv("CERT_KEY_PATH"), r))
	return
}

func DefineRoutes(r *chi.Mux, dashManager *dashboard.Dashboard) {
	// Apply a rate limiter to all routes
	r.Use(httprate.Limit(
		50,             // requests
		60*time.Second, // per durations
		httprate.WithKeyFuncs(httprate.KeyByIP, httprate.KeyByEndpoint),
	))
}
