package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
)

type Person struct {
	ID         primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	FullName   string             `json:"fullname" bson:"fullname"`
	BirthDate  string             `json:"birthDate" bson:"birthDate"`
	Address    Address            `json:"address" bson:"address"`
	Contacts   string             `json:"contacts" bson:"contacts"`
	IsStudent  bool               `json:"isStudent" bson:"isStudent"`
	IsEmployed bool               `json:"isEmployed" bson:"isEmployed"`
}

type Address struct {
	City    string `json:"city" bson:"city"`
	Country string `json:"country" bson:"country"`
}

type JsonResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	people    []Person
	peopleMux sync.Mutex
)

func createUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid method")
		return
	}

	err := r.ParseForm()
	if err != nil {
		fmt.Println("Error parsing form:", err)
		respondWithError(w, http.StatusBadRequest, "Error parsing form")
		return
	}

	fmt.Println("Raw Form Data:", r.Form)

	// Debug: Print the raw request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	fmt.Println("Raw Request Body:", string(body))

	var req Person

	req.FullName = r.Form.Get("fullname")
	req.BirthDate = r.Form.Get("birthDate")
	req.Address.City = r.Form.Get("city")
	req.Address.Country = r.Form.Get("country")
	req.Contacts = r.Form.Get("contacts")
	req.IsStudent, _ = strconv.ParseBool(r.Form.Get("isStudent"))
	req.IsEmployed, _ = strconv.ParseBool(r.Form.Get("isEmployed"))

	// Store the submitted data
	peopleMux.Lock()
	people = append(people, req)
	peopleMux.Unlock()

	// Insert the new user into MongoDB
	result, err := database.Collection("users").InsertOne(r.Context(), req)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Respond with JSON indicating success and the inserted user's ID
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "User successfully created",
		"user_id": result.InsertedID,
	})
}

func createUserThunder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Person
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Process the received person data as needed

	// Store the submitted data
	peopleMux.Lock()
	people = append(people, req)
	peopleMux.Unlock()

	// Insert the new user into MongoDB
	result, err := database.Collection("users").InsertOne(r.Context(), req)
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Respond with JSON indicating success and the inserted user's ID
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "User successfully created",
		"user_id": result.InsertedID,
	})
}

func findUserById(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	log.Println("Searching for user with ID:", userID)

	// Validate if the ID is a valid MongoDB ObjectId
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println("Invalid user ID format:", err)
		respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid user ID format",
		})
		return
	}

	var user Person
	err = database.Collection("users").FindOne(r.Context(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		log.Println("Error:", err)
		respondWithJSON(w, http.StatusNotFound, map[string]interface{}{
			"status":  "error",
			"message": "User not found",
		})
		return
	}

	log.Println("Found user:", user)

}

func updateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Only PUT requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	var updatedUser Person
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	result, err := database.Collection("users").UpdateOne(r.Context(), bson.M{"_id": updatedUser.ID}, bson.M{"$set": updatedUser})
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("User updated: %d document(s) modified", result.ModifiedCount),
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Only DELETE requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	log.Println("Deleting user with ID:", userID)

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println("Invalid user ID format:", err)
		respondWithJSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"message": "Invalid user ID format",
		})
		return
	}

	result, err := database.Collection("users").DeleteOne(r.Context(), bson.M{"_id": objID})
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"status":  "success",
			"message": fmt.Sprintf("User not found: %d document(s) deleted", result.DeletedCount),
		})
		return
	}

}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	cursor, err := database.Collection("users").Find(r.Context(), bson.D{})
	if err != nil {
		log.Fatal(err)
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(r.Context())

	var users []Person
	if err := cursor.All(r.Context(), &users); err != nil {
		log.Fatal(err)
		http.Error(w, "Error decoding users", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"users":  users,
	})
}

func getAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch all users from your existing getAllUsers function
	getAllUsers(w, r)
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	res := JsonResponse{
		Status:  fmt.Sprintf("%d", statusCode),
		Message: message,
	}

	respondWithJSON(w, statusCode, res)
}

func main() {
	// Replace <your_connection_string> with the actual connection string from MongoDB Atlas
	connectionString := "mongodb+srv://aralbaevrasul75:1212@cluster0.iheepym.mongodb.net/?retryWrites=true&w=majority"

	// Set up MongoDB client
	clientOptions := options.Client().ApplyURI(connectionString)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Set up database and collection
	database = client.Database("project3")

	http.HandleFunc("/createUser", createUser)
	http.HandleFunc("/createUserThunder", createUserThunder)
	http.HandleFunc("/findUserById", findUserById)
	http.HandleFunc("/updateUser", updateUser)
	http.HandleFunc("/deleteUser", deleteUser)
	http.HandleFunc("/getAllUsers", getAllUsers)
	http.HandleFunc("/getAllUsersHandler", getAllUsersHandler)
	http.Handle("/", http.FileServer(http.Dir("."))) // Serve files in the current directory

	fmt.Println("Server is running on :8080...")
	http.ListenAndServe(":8080", nil)
}
