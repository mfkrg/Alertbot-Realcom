package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
)

func sendMessage(text string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	params := map[string]string{
		"chat_id":    channelID,
		"text":       text,
		"parse_mode": "Markdown",
	}

	jsonData, _ := json.Marshal(params)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending Telegram message: %v", err)
		return
	}
	defer resp.Body.Close()
}

func sendFile(filePath string) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendDocument", botToken)

	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("Error opening file %s: %v", filePath, err)
		return
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("document", filePath)
	if err != nil {
		log.Printf("Error creating form file: %v", err)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Printf("Error copying file data: %v", err)
		return
	}

	err = writer.WriteField("chat_id", channelID)
	if err != nil {
		log.Printf("Error writing chat_id field: %v", err)
		return
	}

	writer.Close()

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		log.Printf("Error creating Telegram file request: %v", err)
		return
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending file to Telegram: %v", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("Telegram file upload response: %s", string(respBody))
}
