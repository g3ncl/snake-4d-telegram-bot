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
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// ScoreRequest represents the incoming request payload.
// Either InlineMessageID or (ChatID + MessageID) must be provided.
type ScoreRequest struct {
	UserID          string `json:"userId"`
	Score           int    `json:"score"`
	InlineMessageID string `json:"inlineMessageId,omitempty"`
	ChatID          string `json:"chatId,omitempty"`
	MessageID       string `json:"messageId,omitempty"`
}

// InlineTelegramRequest is the payload for inline mode setGameScore
type InlineTelegramRequest struct {
	InlineMessageID string `json:"inline_message_id"`
	UserID          int64  `json:"user_id"`
	Score           int    `json:"score"`
}

// DirectTelegramRequest is the payload for direct mode setGameScore
type DirectTelegramRequest struct {
	ChatID    int64 `json:"chat_id"`
	MessageID int   `json:"message_id"`
	UserID    int64 `json:"user_id"`
	Score     int   `json:"score"`
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

	log.Printf("Parsed request: UserID=%q, Score=%d, InlineMessageID=%q, ChatID=%q, MessageID=%q",
		scoreReq.UserID, scoreReq.Score, scoreReq.InlineMessageID, scoreReq.ChatID, scoreReq.MessageID)

	// Validate required fields
	if scoreReq.UserID == "" || scoreReq.Score == 0 {
		log.Printf("Missing fields: UserID=%q, Score=%d", scoreReq.UserID, scoreReq.Score)
		return errorResponse(headers, 400, "Missing required fields")
	}

	// Validate that either inline or direct mode params are provided
	hasInline := scoreReq.InlineMessageID != ""
	hasDirect := scoreReq.ChatID != "" && scoreReq.MessageID != ""
	if !hasInline && !hasDirect {
		log.Printf("Missing message identification: need either inlineMessageId or chatId+messageId")
		return errorResponse(headers, 400, "Missing message identification fields")
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

	userID, err := strconv.ParseInt(scoreReq.UserID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID %q: %w", scoreReq.UserID, err)
	}

	var jsonData []byte

	if scoreReq.InlineMessageID != "" {
		// Inline mode
		telegramReq := InlineTelegramRequest{
			InlineMessageID: scoreReq.InlineMessageID,
			UserID:          userID,
			Score:           scoreReq.Score,
		}
		jsonData, err = json.Marshal(telegramReq)
	} else {
		// Direct mode
		chatID, chatErr := strconv.ParseInt(scoreReq.ChatID, 10, 64)
		if chatErr != nil {
			return fmt.Errorf("invalid chat ID %q: %w", scoreReq.ChatID, chatErr)
		}
		messageID, msgErr := strconv.Atoi(scoreReq.MessageID)
		if msgErr != nil {
			return fmt.Errorf("invalid message ID %q: %w", scoreReq.MessageID, msgErr)
		}
		telegramReq := DirectTelegramRequest{
			ChatID:    chatID,
			MessageID: messageID,
			UserID:    userID,
			Score:     scoreReq.Score,
		}
		jsonData, err = json.Marshal(telegramReq)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal telegram request: %w", err)
	}

	log.Printf("Sending to Telegram API: %s", string(jsonData))

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
