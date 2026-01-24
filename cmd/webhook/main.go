package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	gameShortName = "snake4d"
)

// Handler is the Lambda function handler for Telegram webhook
func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	if request.Body == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid request body",
		}, nil
	}

	// Get Telegram bot token from environment
	botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if botToken == "" {
		log.Println("TELEGRAM_BOT_TOKEN is not set")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Server configuration error",
		}, nil
	}

	// Get game URL from environment
	gameURL := os.Getenv("GAME_URL")
	if gameURL == "" {
		log.Println("GAME_URL is not set")
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Server configuration error",
		}, nil
	}

	// Create bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Printf("Failed to create bot: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Failed to initialize bot",
		}, nil
	}

	// Parse the update
	var update tgbotapi.Update
	if err := json.Unmarshal([]byte(request.Body), &update); err != nil {
		log.Printf("Failed to parse update: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Invalid update format",
		}, nil
	}

	// Handle different update types
	if err := handleUpdate(bot, &update, gameURL); err != nil {
		log.Printf("Error handling update: %v", err)
		// Don't return error to Telegram, just log it
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "OK",
	}, nil
}

// handleUpdate processes different types of Telegram updates
func handleUpdate(bot *tgbotapi.BotAPI, update *tgbotapi.Update, gameURL string) error {
	// Handle text messages
	if update.Message != nil && update.Message.Text != "" {
		return handleMessage(bot, update.Message)
	}

	// Handle inline queries
	if update.InlineQuery != nil {
		return handleInlineQuery(bot, update.InlineQuery)
	}

	// Handle callback queries (game button clicks)
	if update.CallbackQuery != nil && update.CallbackQuery.GameShortName == gameShortName {
		return handleCallbackQuery(bot, update.CallbackQuery, gameURL)
	}

	return nil
}

// handleMessage processes text messages
func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) error {
	switch message.Text {
	case "/start":
		msg := tgbotapi.NewMessage(
			message.Chat.ID,
			"Welcome to the Snake 4D Game Bot! Use @snake4dbot followed by some text in any chat to start playing.",
		)
		_, err := bot.Send(msg)
		return err

	case "/game":
		game := tgbotapi.GameConfig{
			BaseChat: tgbotapi.BaseChat{
				ChatID: message.Chat.ID,
			},
			GameShortName: gameShortName,
		}
		_, err := bot.Send(game)
		return err
	}

	return nil
}

// handleInlineQuery processes inline queries
func handleInlineQuery(bot *tgbotapi.BotAPI, query *tgbotapi.InlineQuery) error {
	result := tgbotapi.InlineQueryResultGame{
		Type: "game",
		ID:   gameShortName,
		GameShortName: gameShortName,
	}
	config := tgbotapi.InlineConfig{
		InlineQueryID: query.ID,
		Results:       []interface{}{result},
	}

	_, err := bot.Request(config)
	return err
}

// handleCallbackQuery processes callback queries (game start)
func handleCallbackQuery(bot *tgbotapi.BotAPI, query *tgbotapi.CallbackQuery, gameURL string) error {
	// Build game URL with user ID and message ID
	url := fmt.Sprintf(
		"%s#userId=%d&messageId=%s",
		gameURL,
		query.From.ID,
		query.InlineMessageID,
	)

	config := tgbotapi.NewCallbackWithAlert(query.ID, "")
	config.URL = url

	_, err := bot.Request(config)
	return err
}

func main() {
	lambda.Start(Handler)
}
