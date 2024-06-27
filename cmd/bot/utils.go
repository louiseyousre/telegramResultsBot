package main

import (
	"bytes"
	"fmt"
	"github.com/go-telegram/bot/models"
	"io"
	"net/http"
)

func replyParametersTo(msg *models.Message) *models.ReplyParameters {
	if msg == nil {
		return nil
	}

	return &models.ReplyParameters{
		MessageID:                msg.ID,
		ChatID:                   msg.Chat.ID,
		AllowSendingWithoutReply: true,
	}
}

func getStudentPhoto(imagePath string) (*models.InputFileUpload, error) {
	// Make a GET request to download the image
	resp, err := http.Get("http://stda.minia.edu.eg" + imagePath)
	if err != nil {
		return nil, fmt.Errorf("error downloading image, %v\n", err)
	}
	defer func(resp *http.Response) {
		_ = resp.Body.Close()
	}(resp)

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error: received non-200 response code %d\n", resp.StatusCode)
	}

	// Read the image data from the response body
	var imageData []byte
	imageData, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading image data, %v\n", err)
	}

	// Create a new reader for the image data
	imageReader := bytes.NewReader(imageData)

	// Prepare the parameters for the photo
	return &models.InputFileUpload{Filename: "student_photo.jpg", Data: imageReader}, nil
}
