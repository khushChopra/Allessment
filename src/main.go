package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// Description to image ID map asdasd asds
var globalMap map[string]string

func ConverseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body
	var requestBody map[string]interface{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Implement your logic to process the conversation and set the response data
	// For example:
	responseData := map[string]interface{}{"response": "This is the response to the conversation."}

	// Set the response header and send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseData)
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {

	newUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		return
	}
	// Take image and string
	// Update map and store image
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err = r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Get the description from the form
	description := r.FormValue("description")

	// Handle the uploaded file (image)
	file, handler, err := r.FormFile("image")
	// fmt.Println(handler.Filename)
	filename := newUUID.String() + "-" + handler.Filename
	if err != nil {
		http.Error(w, "Unable to get the image file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	destinationFile, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating destination file:", err)
		return
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, file)
	if err != nil {
		http.Error(w, "Unable to copy the file", http.StatusInternalServerError)
		return
	}
	globalMap[description] = filename
	// Implement your logic to save the image with a unique ID and store the mapping
	// For example, save the image to disk with a unique ID and store the mapping in memory.

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Image uploaded with description: %s", description)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the description query parameter
	description := r.URL.Query().Get("description")

	// Implement your logic to find and return the best matching image based on the description
	// For example, retrieve the image based on the description.
	for key, value := range globalMap {
		fmt.Printf("Key: %s, Value: %d\n", key, value)
	}
	imagePath, exists := globalMap[description]

	if exists {
		imageFile, err := os.Open(imagePath)
		if err != nil {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		defer imageFile.Close()
		// Set the response header and send the image as a download
		w.Header().Set("Content-Disposition", "attachment; filename="+imagePath)
		rtype := "image/png"
		// if !strings.containes(imagePath, ".png") {
		// 	rtype = "image/jpg"
		// }
		w.Header().Set("Content-Type", rtype)
		io.Copy(w, imageFile)
		// io.WriteString(w, "Binary image data goes here") // Replace with your image data
	} else {
		http.Error(w, "No file found", http.StatusBadRequest)
	}

}

func main() {
	globalMap = make(map[string]string)
	// http.HandleFunc("/converse", ConverseHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/download", DownloadHandler)

	port := 8080
	fmt.Printf("Starting server on :%d...\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
