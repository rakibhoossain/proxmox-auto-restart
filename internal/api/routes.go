package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// SetupRoutes configures all HTTP routes with middleware
func SetupRoutes() http.Handler {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check (no auth required)
	r.Get("/health", HealthCheck)

	// API routes (with authentication)
	r.Route("/api", func(r chi.Router) {
		// Apply basic auth to all API routes
		r.Use(BasicAuthMiddleware)

		// Resources (VMs and Containers)
		r.Route("/resources", func(r chi.Router) {
			r.Get("/", GetResources)                   // GET /api/resources
			r.Get("/{vmid}", GetResource)              // GET /api/resources/103?node=www
			r.Post("/{vmid}/restart", RestartResource) // POST /api/resources/103/restart?node=www
			r.Post("/{vmid}/stop", StopResource)       // POST /api/resources/103/stop?node=www
			r.Post("/{vmid}/start", StartResource)     // POST /api/resources/103/start?node=www
		})

		// Whitelist
		r.Route("/whitelist", func(r chi.Router) {
			r.Get("/", GetWhitelist)               // GET /api/whitelist
			r.Post("/", AddToWhitelist)            // POST /api/whitelist
			r.Put("/{id}", UpdateWhitelist)        // PUT /api/whitelist/1
			r.Delete("/{id}", DeleteFromWhitelist) // DELETE /api/whitelist/1
		})

		// Logs
		r.Route("/logs", func(r chi.Router) {
			r.Get("/", GetLogs) // GET /api/logs?vmid=103&status=success
		})

		// System
		r.Get("/status", GetStatus) // GET /api/status
	})

	return r
}
