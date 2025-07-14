package api

import (
	"encoding/json"
	"net/http"

	"github.com/Karan2980/llm-planner-golang-project/internal/models"
	"github.com/Karan2980/llm-planner-golang-project/internal/planner"
	"github.com/gorilla/mux"
	calendarv3 "google.golang.org/api/calendar/v3"
)

// Server represents the API server
type Server struct {
	scheduler *planner.EnhancedScheduler
	router    *mux.Router
}

// NewServer creates a new API server
func NewServer(calendarService *calendarv3.Service, aiConfig models.AIConfig, timeZone string) *Server {
	scheduler := planner.NewEnhancedScheduler(calendarService, aiConfig, timeZone)
	
	server := &Server{
		scheduler: scheduler,
		router:    mux.NewRouter(),
	}
	
	server.setupRoutes()
	return server
}

// Router returns the HTTP router
func (s *Server) Router() *mux.Router {
	return s.router
}

// setupRoutes sets up all API routes
// setupRoutes sets up all API routes
func (s *Server) setupRoutes() {
	// Add CORS middleware
	s.router.Use(corsMiddleware)
	
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	
	// api.HandleFunc("/schedule", s.handleSchedule).Methods("POST")
	// api.HandleFunc("/query", s.handleQuery).Methods("POST")
	// api.HandleFunc("/events/today", s.handleTodaysEvents).Methods("GET")
	// api.HandleFunc("/events/upcoming", s.handleUpcomingEvents).Methods("GET")
	// api.HandleFunc("/search", s.handleSearch).Methods("POST")
	// api.HandleFunc("/stats", s.handleStats).Methods("GET")
	// api.HandleFunc("/delete", s.handleDelete).Methods("POST")
	
	// KEEP ONLY THIS:
	api.HandleFunc("/unified", s.handleUnifiedQuery).Methods("POST")
	
	// Health check
	s.router.HandleFunc("/health", s.handleHealth).Methods("GET")
	
	// Root endpoint with API documentation
	s.router.HandleFunc("/", s.handleRoot).Methods("GET")
}



// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// writeJSON writes JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes error response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{"error": message})
}
