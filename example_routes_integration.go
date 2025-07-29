// Example of how to integrate the game config routes into your main router
// This would go in your main.go or router setup file

package main

import (
	"github.com/gorilla/mux"
	"github.com/example/officestonks/internal/handlers"
	"github.com/example/officestonks/internal/middleware"
)

func setupGameConfigRoutes(r *mux.Router, gameConfigHandler *handlers.GameConfigHandler) {
	// Admin-only game configuration routes
	adminRouter := r.PathPrefix("/api/admin").Subrouter()
	
	// Apply admin authentication middleware
	adminRouter.Use(middleware.JWTAuth)
	adminRouter.Use(middleware.AdminOnly)
	
	// Game configuration routes
	adminRouter.HandleFunc("/game-config", gameConfigHandler.GetGameConfig).Methods("GET")
	adminRouter.HandleFunc("/game-config", gameConfigHandler.UpdateGameConfig).Methods("PUT")
	adminRouter.HandleFunc("/game-config/reset", gameConfigHandler.ResetGameConfig).Methods("POST")
	adminRouter.HandleFunc("/game-config/balanced", gameConfigHandler.LoadBalancedConfig).Methods("POST")
}

// Example of how to initialize in main.go
func exampleMain() {
	// ... other setup code ...
	
	// Create game config handler
	gameConfigHandler := handlers.NewGameConfigHandler()
	
	// Setup routes
	r := mux.NewRouter()
	setupGameConfigRoutes(r, gameConfigHandler)
	
	// ... rest of server setup ...
}