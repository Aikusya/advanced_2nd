package main

import (
	"advanced_2nd/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
)

var (
	people    []models.Person
	peopleMux sync.Mutex
)

func main() {
	http.Handle("/", http.FileServer(http.Dir("."))) // Serve files in the current directory
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/thunderSubmit", thunderSubmitHandler)
	http.HandleFunc("/getPerson", getPersonHandler)
	fmt.Println("Server is running on :8080...")
	http.ListenAndServe(":8080", nil)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
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

	var req models.Person
	req.ID, _ = strconv.Atoi(r.Form.Get("id"))
	req.FullName = r.Form.Get("fullname")
	req.BirthDate = r.Form.Get("birthDate")
	req.Address.City = r.Form.Get("city")
	req.Address.Country = r.Form.Get("country")
	req.Contacts = append(req.Contacts, r.Form.Get("contacts"))
	req.IsStudent, _ = strconv.ParseBool(r.Form.Get("isStudent"))
	req.IsEmployed, _ = strconv.ParseBool(r.Form.Get("isEmployed"))

	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	fmt.Println("Decoded JSON:", req)

	// Process the received person data as needed

	// Store the submitted data
	peopleMux.Lock()
	defer peopleMux.Unlock()

	people = append(people, req) // Append the new req to the people slice

	res := models.JsonResponse{
		Status:  "success",
		Message: "Data successfully received",
	}

	respondWithJSON(w, http.StatusOK, res)
}

func thunderSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid method")
		return
	}

	var req models.Person
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		respondWithError(w, http.StatusBadRequest, "Invalid JSON format")
		return
	}
	fmt.Println("Received JSON:", req)

	// Process the received person data as needed

	// Store the submitted data
	peopleMux.Lock()
	people = append(people, req)
	peopleMux.Unlock()

	res := models.JsonResponse{
		Status:  "success",
		Message: "Data successfully received",
	}

	respondWithJSON(w, http.StatusOK, res)
}

func getPersonHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Invalid method")
		return
	}

	peopleMux.Lock()
	defer peopleMux.Unlock()

	// Print the number of people in the stored data (for debugging purposes)
	fmt.Println("Number of people:", len(people))

	// Return the last person in the stored data
	if len(people) > 0 {
		respondWithJSON(w, http.StatusOK, people[len(people)-1])
		return
	}

	respondWithError(w, http.StatusNotFound, "No submitted data found")
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	res := models.JsonResponse{
		Status:  fmt.Sprintf("%d", statusCode),
		Message: message,
	}

	respondWithJSON(w, statusCode, res)
}
