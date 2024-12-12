package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Dish struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"`
	History      string   `json:"history"`
	Ingredients  []string `json:"ingredients"`
	Recipe       string   `json:"recipe"`
	Instructions []string `json:"instructions"`
	PictureURL   string   `json:"picture_url"`
}

var dishes []Dish

func main() {
    // Load dishes from JSON file
    if err := loadDishes("dishes.json"); err != nil {
        fmt.Printf("Error loading dishes: %v\n", err)
        os.Exit(1)
    }

    // Apply the CORS middleware to all routes
    http.HandleFunc("/api/dishes", corsMiddleware(handleDishes))    // GET all dishes, POST new dish
    http.HandleFunc("/api/dishes/", corsMiddleware(handleDishByID)) // GET, PUT, DELETE dish by ID

    // Start the server
    fmt.Println("Server running on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
}

func loadDishes(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &dishes)
}

func saveDishes(filePath string) error {
	data, err := json.MarshalIndent(dishes, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, data, 0644)
}

func getNextID() int {
	maxID := 0
	for _, dish := range dishes {
		if dish.ID > maxID {
			maxID = dish.ID
		}
	}
	return maxID + 1
}

// CORS Middleware
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set the correct allowed origin here
        w.Header().Set("Access-Control-Allow-Origin", "*")

		// Set the allowed methods and headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests (OPTIONS requests)
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Continue processing the request
		next(w, r)
	}
}




func handleDishes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Get all dishes
		json.NewEncoder(w).Encode(dishes)

	case "POST":
		// Create a new dish
		var newDish Dish
		if err := json.NewDecoder(r.Body).Decode(&newDish); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		newDish.ID = getNextID()
		dishes = append(dishes, newDish)
		if err := saveDishes("dishes.json"); err != nil {
			http.Error(w, "Failed to save data", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newDish)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleDishByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract the ID from the URL
	idStr := strings.TrimPrefix(r.URL.Path, "/api/dishes/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid dish ID", http.StatusBadRequest)
		return
	}

	// Find the dish by ID
	var dish *Dish
	for i, d := range dishes {
		if d.ID == id {
			dish = &dishes[i]
			break
		}
	}

	if dish == nil {
		http.Error(w, "Dish not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		// Get the dish by ID
		json.NewEncoder(w).Encode(dish)

	case "PUT":
		// Update the dish
		var updatedDish Dish
		if err := json.NewDecoder(r.Body).Decode(&updatedDish); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}
		dish.Name = updatedDish.Name
		dish.History = updatedDish.History
		dish.Ingredients = updatedDish.Ingredients
		dish.Recipe = updatedDish.Recipe
		dish.Instructions = updatedDish.Instructions
		dish.PictureURL = updatedDish.PictureURL

		if err := saveDishes("dishes.json"); err != nil {
			http.Error(w, "Failed to save data", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dish)

	case "DELETE":
		// Delete the dish
		for i, d := range dishes {
			if d.ID == id {
				dishes = append(dishes[:i], dishes[i+1:]...)
				if err := saveDishes("dishes.json"); err != nil {
					http.Error(w, "Failed to save data", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		http.Error(w, "Dish not found", http.StatusNotFound)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
