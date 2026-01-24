package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// ScoreRequest represents the incoming request payload
type ScoreRequest struct {
	UserID    string `json:"userId"`
	Score     int    `json:"score"`
	MessageID string `json:"messageId"`
}

// TelegramRequest represents the payload sent to Telegram API
type TelegramRequest struct {
	InlineMessageID string `json:"inline_message_id"`
	UserID          string `json:"user_id"`
	Score           int    `json:"score"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Success bool `json:"success"`
}

// Handler is the Lambda function handler
func Handler(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request. Method: %s. Body length: %d", request.RequestContext.HTTP.Method, len(request.Body))

	// CORS headers for all responses
	headers := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Headers": "Content-Type",
		"Access-Control-Allow-Methods": "OPTIONS,POST",
		"Content-Type":                 "application/json",
	}

	// Handle preflight OPTIONS request
	if request.RequestContext.HTTP.Method == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers:    headers,
		}, nil
	}

	// Get Telegram bot token from environment
	botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if botToken == "" {
		log.Println("Error: TELEGRAM_BOT_TOKEN is not set")
		return errorResponse(headers, 500, "Server configuration error")
	}

	// Parse request body
	var scoreReq ScoreRequest
	if err := json.Unmarshal([]byte(request.Body), &scoreReq); err != nil {
		log.Printf("Error decoding body: %v. Body header: %q", err, string(request.Body))
		return errorResponse(headers, 400, "Invalid request body")
	}
	
	log.Printf("Parsed request: UserID=%q, Score=%d, MessageID=%q", scoreReq.UserID, scoreReq.Score, scoreReq.MessageID)

	// Validate required fields
	if scoreReq.UserID == "" || scoreReq.Score == 0 || scoreReq.MessageID == "" {
		log.Printf("Missing fields: UserID=%q, Score=%d, MessageID=%q", scoreReq.UserID, scoreReq.Score, scoreReq.MessageID)
		return errorResponse(headers, 400, "Missing required fields")
	}

	// Call Telegram API to update score
	if err := updateTelegramScore(botToken, scoreReq); err != nil {
		log.Printf("Error updating score: %v", err)
		return errorResponse(headers, 500, fmt.Sprintf("Failed to update score: %v", err))
	}

	log.Printf("Successfully updated score for UserID=%s", scoreReq.UserID)

	// Return success response
	successBody, _ := json.Marshal(SuccessResponse{Success: true})
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       string(successBody),
	}, nil
}

// updateTelegramScore sends the score update to Telegram Bot API
func updateTelegramScore(botToken string, scoreReq ScoreRequest) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/setGameScore", botToken)

	telegramReq := TelegramRequest{
		InlineMessageID: scoreReq.MessageID,
		UserID:          scoreReq.UserID,
		Score:           scoreReq.Score,
	}

	jsonData, err := json.Marshal(telegramReq)
	if err != nil {
		return fmt.Errorf("failed to marshal telegram request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to call telegram API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// errorResponse creates an error response with proper headers
func errorResponse(headers map[string]string, statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	errorBody, _ := json.Marshal(ErrorResponse{
		Success: false,
		Error:   message,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       string(errorBody),
	}, nil
}

func main() {
	lambda.Start(Handler)
}
