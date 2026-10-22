package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore" // <--- REQUIRED FOR FIRESTORE
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Structs representing data models
type Todo struct {
	ID     string `json:"id"` // Firestore uses Document IDs (Strings)
	Text   string `json:"text"`
	Date   string `json:"date"`
	Time   string `json:"time"`
	Done   bool   `json:"done"`
	IsLate bool   `json:"is_late"` // Calculated dynamically
}

var db *firestore.Client // <--- CHANGED TYPE TO FIRESTORE CLIENT
var ctx context.Context

func main() {
	ctx = context.Background()

	// Initialize Firebase App with Service Account Key
	sa := option.WithCredentialsFile("serviceAccountKey.json")
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	// Get Firestore Client (NOT Database/RTDB)
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Error creating Firestore client: %v", err)
	}
	defer client.Close()

	db = client
	log.Println("Connected to Cloud Firestore successfully")

	// Define API routes
	mux := http.NewServeMux()

	mux.HandleFunc("/api/todos", handleGetTodos)
	mux.HandleFunc("/api/todo/add", handleAddTodo)
	mux.HandleFunc("/api/todo/update", handleUpdateTodo)
	mux.HandleFunc("/api/todo/delete", handleDeleteTodo)
	mux.HandleFunc("/api/todo/toggle", handleToggleTodo)

	handler := corsMiddleware(mux) // <--- Ensure mux is passed here

	port := ":8083"
	log.Printf("API Server running on http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS") // Added PATCH explicitly
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- FIRESTORE HANDLERS ---

const collectionName = "todos"

func handleGetTodos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	iter := db.Collection(collectionName).OrderBy("date", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	var todos []Todo
	now := time.Now()
	todayStr := now.Format("2006-01-02")
	currentTimeStr := now.Format("15:04")

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var t Todo
		err = doc.DataTo(&t)
		if err != nil {
			continue
		}
		t.ID = doc.Ref.ID // Set the ID from the document reference

		// Calculate IsLate logic based on current time
		isPastDate := t.Date < todayStr
		isTodayButLate := (t.Date == todayStr && t.Time < currentTimeStr)
		t.IsLate = !t.Done && (isPastDate || isTodayButLate)

		todos = append(todos, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(todos)
}

func handleAddTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		Text string `json:"text"`
		Date string `json:"date"`
		Time string `json:"time"`
	}

	var req RequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Add document to Firestore
	_, _, err = db.Collection(collectionName).Add(ctx, map[string]interface{}{
		"text": req.Text,
		"date": req.Date,
		"time": req.Time,
		"done": false,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task added"})
}

func handleUpdateTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		ID   string `json:"id"` // Note: ID is now a String
		Text string `json:"text"`
		Date string `json:"date"`
		Time string `json:"time"`
	}

	var req RequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Collection(collectionName).Doc(req.ID).Update(ctx, []firestore.Update{
		{Path: "text", Value: req.Text},
		{Path: "date", Value: req.Date},
		{Path: "time", Value: req.Time},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task updated"})
}

func handleDeleteTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		ID string `json:"id"` // Note: ID is now a String
	}

	var req RequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Collection(collectionName).Doc(req.ID).Delete(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Task deleted"})
}

// FIXED TOGGLE FUNCTION WITH BETTER ERROR LOGGING
func handleToggleTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		ID string `json:"id"`
	}

	var req RequestBody
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to toggle task ID: %s", req.ID)

	// 1. Get the document reference
	docRef := db.Collection(collectionName).Doc(req.ID)

	// 2. Fetch current data safely
	snap, err := docRef.Get(ctx)
	if err != nil {
		log.Printf("Error fetching doc %s: %v", req.ID, err)
		http.Error(w, "Document not found or error fetching", http.StatusInternalServerError)
		return
	}

	if !snap.Exists() {
		log.Printf("Document %s does not exist in Firestore", req.ID)
		http.Error(w, "Document does not exist", http.StatusNotFound)
		return
	}

	// 3. Extract 'done' field carefully
	data := snap.Data()
	currentDone, ok := data["done"].(bool)
	if !ok {
		// If 'done' key doesn't exist or isn't boolean, default to false
		currentDone = false
	}

	// 4. Flip the status
	newStatus := !currentDone
	log.Printf("Toggling ID %s from %v to %v", req.ID, currentDone, newStatus)

	// 5. Update Firestore
	_, err = docRef.Update(ctx, []firestore.Update{
		{Path: "done", Value: newStatus},
	})

	if err != nil {
		log.Printf("ERROR UPDATING DOC %s: %v", req.ID, err)
		http.Error(w, "Failed to update status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Status toggled"})
}
