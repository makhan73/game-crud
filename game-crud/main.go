package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

var ctx = context.Background()
var rdb *redis.Client

// Game represents the structure of a game
type Game struct {
	GameID      string `json:"game_id"`
	Game        string `json:"game"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Created     string `json:"created"`
	Deleted     bool   `json:"deleted"`
}

func init() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func main() {
	r := mux.NewRouter()

	// Serve static files (UI)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/")))

	// API Routes
	r.HandleFunc("/games", createGame).Methods("POST")
	r.HandleFunc("/games/{game_id}", getGame).Methods("GET")
	r.HandleFunc("/games/{game_id}", updateGame).Methods("PUT")
	r.HandleFunc("/games/{game_id}", deleteGame).Methods("DELETE")
	r.HandleFunc("/games", listGames).Methods("GET")

	fmt.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

// Create a new game
func createGame(w http.ResponseWriter, r *http.Request) {
	var game Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	game.Created = time.Now().Format(time.RFC3339)
	game.Deleted = false

	gameJSON, err := json.Marshal(game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rdb.Set(ctx, game.GameID, gameJSON, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

// Get a game by ID
func getGame(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	gameID := params["game_id"]

	val, err := rdb.Get(ctx, gameID).Result()
	if err == redis.Nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var game Game
	json.Unmarshal([]byte(val), &game)
	json.NewEncoder(w).Encode(game)
}

// Update a game by ID
func updateGame(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	gameID := params["game_id"]

	var updatedGame Game
	if err := json.NewDecoder(r.Body).Decode(&updatedGame); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	val, err := rdb.Get(ctx, gameID).Result()
	if err == redis.Nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var existingGame Game
	json.Unmarshal([]byte(val), &existingGame)

	// Update fields
	if updatedGame.Game != "" {
		existingGame.Game = updatedGame.Game
	}
	if updatedGame.Description != "" {
		existingGame.Description = updatedGame.Description
	}
	if updatedGame.Status != "" {
		existingGame.Status = updatedGame.Status
	}

	gameJSON, err := json.Marshal(existingGame)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rdb.Set(ctx, gameID, gameJSON, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingGame)
}

// Delete a game by ID (soft delete)
func deleteGame(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	gameID := params["game_id"]

	val, err := rdb.Get(ctx, gameID).Result()
	if err == redis.Nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var game Game
	json.Unmarshal([]byte(val), &game)
	game.Deleted = true

	gameJSON, err := json.Marshal(game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = rdb.Set(ctx, gameID, gameJSON, 0).Err()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(game)
}

// List all games
func listGames(w http.ResponseWriter, r *http.Request) {
	keys, err := rdb.Keys(ctx, "*").Result()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var games []Game
	for _, key := range keys {
		val, err := rdb.Get(ctx, key).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var game Game
		json.Unmarshal([]byte(val), &game)
		if !game.Deleted {
			games = append(games, game)
		}
	}

	json.NewEncoder(w).Encode(games)
}
