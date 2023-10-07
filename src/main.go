package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"github.com/google/uuid"
	"context"
	openai "github.com/sashabaranov/go-openai"
)

// Description to image ID map
var globalMap map[string]string
var OPENAI_KEY string

type Message struct {
    Role string `json:"role"`
    Msg  string `json:"msg"`
}

type Request struct {
    Msg     string         `json:"msg"`
    History []Message `json:"history"`
}

func GetGPTResponse(req Request) (string, error) {
	messages := make([]openai.ChatCompletionMessage, 0)
	for _, v := range req.History {
		role := v.Role
		if role!="assistant" {
			role = "user"
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: v.Msg,
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    "user",
		Content: req.Msg,
	})

	client := openai.NewClient(OPENAI_KEY)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: messages,
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	msg := resp.Choices[0].Message.Content
	fmt.Println(msg)
	return msg, nil
}

func ImageIntentChecker(msg string) (string, error){
	client := openai.NewClient(OPENAI_KEY)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:	openai.ChatMessageRoleSystem,
					Content: "You need to tell if this person is trying to hold a normal conversaion or is he trying to uplaod or download an image or picture. If he wants to download, say 'download_image'. If he wants to upload, say 'upload_image'. Else say 'None'",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: msg,
				},
			},
		},
	)
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}
	fmt.Println(resp.Choices[0].Message.Content)
	return resp.Choices[0].Message.Content, nil
}

func ConverseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	intent, err := ImageIntentChecker(req.Msg)

	if err!=nil {
		http.Error(w, "Some issue with OpenAI", http.StatusInternalServerError)
		return
	}
	if intent=="download_image" {
		responseData := map[string]interface{}{"intent": "download_image", "msg": ""}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseData)
	} else if intent=="upload_image" {
		responseData := map[string]interface{}{"intent": "upload_image", "msg": ""}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseData)
	} else {
		res, err := GetGPTResponse(req)
		if err!=nil {
			http.Error(w, "Some issue with OpenAI", http.StatusInternalServerError)
			return
		}
		responseData := map[string]interface{}{"intent": "", "msg": res}
		// Set the response header and send the JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(responseData)
	}
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	newUUID, err := uuid.NewRandom()
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	err = r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}
	description := r.FormValue("description")
	file, handler, err := r.FormFile("image")
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
	w.WriteHeader(http.StatusOK)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	description := r.URL.Query().Get("description")
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
		w.Header().Set("Content-Disposition", "attachment; filename="+imagePath)
		rtype := "image/png"
		w.Header().Set("Content-Type", rtype)
		io.Copy(w, imageFile)
	} else {
		http.Error(w, "No file found", http.StatusBadRequest)
		return 
	}
}

func ImageListHandler(w http.ResponseWriter, r *http.Request) {
	strings := make([]string, 0)
	i := 0
	for key, _ := range globalMap {
		strings = append(strings, key)
		i = i+1
		if i>5 {
			break
		}
	}
	jsonData, err := json.Marshal(strings)
	if err != nil {
		http.Error(w, "JSON encoding error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	OPENAI_KEY = os.Getenv("OPENAI_KEY")
	globalMap = make(map[string]string)
	http.HandleFunc("/converse", ConverseHandler)
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/download", DownloadHandler)
	http.HandleFunc("/list", ImageListHandler)
	port := 8080
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}