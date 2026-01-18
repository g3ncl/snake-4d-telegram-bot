# Snake 4D Telegram Bot

This repository contains the code for [@snake4dbot](https://t.me/snake4dbot), a Telegram bot that allows users to play the Snake 4D game directly within Telegram chats. The bot is built using Go and is designed to be deployed on AWS Lambda.

## Features

- Responds to `/start` command with a welcome message
- Responds to `/game` command by sending the Snake 4D game
- Supports inline queries to start the game in any chat
- Integrates with the Snake 4D game hosted at https://snake4d.netlify.app
- Utilizes Telegram's Game API to set and display highscores within chats
- Provides a dedicated score update endpoint for the game frontend

## Architecture

This repository deploys **two separate Lambda functions**, both written in Go:

1. **`snake-4d-telegram-bot`** (Webhook Handler): Handles Telegram webhook events (commands, inline queries, callback queries)
2. **`snake-4d-score-update`** (Score Update API): Handles score update requests from the game frontend via a Function URL

## Prerequisites

- Go (v1.21 or later)
- AWS account with Lambda and IAM access
- Telegram Bot Token

## Setup

1. Clone this repository:
   ```bash
   git clone https://github.com/g3ncl/snake-4d-telegram-bot.git
   cd snake-4d-telegram-bot
   ```

2. Install Go dependencies:
   ```bash
   go mod download
   ```

3. Create a `.env` file in the root directory and add your Telegram Bot Token:
   ```
   TELEGRAM_BOT_TOKEN=your_bot_token
   ```

4. Build the Go Lambda functions:
   ```bash
   mkdir -p bin/webhook bin/score-update
   GOOS=linux GOARCH=arm64 go build -o bin/webhook/bootstrap cmd/webhook/main.go
   GOOS=linux GOARCH=arm64 go build -o bin/score-update/bootstrap cmd/score-update/main.go
   cd bin/webhook && zip ../../webhook.zip bootstrap && cd ../..
   cd bin/score-update && zip ../../score-update.zip bootstrap && cd ../..
   ```

## Deployment

This bot is configured to be automatically deployed to AWS Lambda using GitHub Actions. The workflow is triggered on every push to the `main` branch and deploys both Lambda functions.

### Authentication

The workflow uses **OIDC (OpenID Connect)** to authenticate with AWS, eliminating the need for long-lived access keys. It assumes the existing `GitHubActionsDeployRole` IAM role.

### Required GitHub Secrets

Set up the following secrets in your GitHub repository (Settings → Secrets and variables → Actions):

- `AWS_REGION`: The AWS region where your Lambda functions are located (e.g., `eu-south-1`)
- `TELEGRAM_BOT_TOKEN`: Your Telegram Bot Token
- `GAME_URL`: The URL where the Snake 4D game is hosted (e.g., `https://snake4d.netlify.app`)
- `LAMBDA_EXECUTION_ROLE`: The ARN of the IAM role for Lambda execution

### Function URL Setup (First-Time Only)

After the first deployment of the `snake-4d-score-update` Lambda function, you need to create a Function URL:

1. Navigate to the AWS Lambda console
2. Select the `snake-4d-score-update` function
3. Go to **Configuration** → **Function URL**
4. Click **Create function URL**
   - **Auth type**: `NONE` (public access)
   - **Configure cross-origin resource sharing (CORS)**: Optional (handler includes CORS headers)
5. Copy the generated Function URL (e.g., `https://abc123.lambda-url.eu-south-1.on.aws/`)
6. Add this URL to your frontend's GitHub secrets as `NEXT_PUBLIC_SCORE_API_URL`

### Manual Deployment

If you prefer to deploy manually or need to set up the Lambda functions for the first time:

**Build both Lambda functions:**

```bash
mkdir -p bin/webhook bin/score-update
GOOS=linux GOARCH=arm64 go build -o bin/webhook/bootstrap cmd/webhook/main.go
GOOS=linux GOARCH=arm64 go build -o bin/score-update/bootstrap cmd/score-update/main.go
cd bin/webhook && zip ../../webhook.zip bootstrap && cd ../..
cd bin/score-update && zip ../../score-update.zip bootstrap && cd ../..
```

**Deploy the webhook handler:**

Use the AWS CLI or AWS Management Console to create or update the `snake-4d-telegram-bot` Lambda function:
- Upload `webhook.zip`
- Set runtime to `provided.al2023` with `arm64` architecture
- Set handler to `bootstrap`
- Set environment variables:
  - `TELEGRAM_BOT_TOKEN`: Your Telegram bot token
  - `GAME_URL`: The URL where the Snake 4D game is hosted (e.g., `https://snake4d.netlify.app`)
- Set up an API Gateway trigger and configure the Telegram webhook URL

**Deploy the score update handler:**

Use the AWS CLI or AWS Management Console to create or update the `snake-4d-score-update` Lambda function:
   - Upload `score-update.zip`
   - Set runtime to `provided.al2023` with `arm64` architecture
   - Set handler to `bootstrap`
   - Set environment variable `TELEGRAM_BOT_TOKEN`
   - Create a Function URL (see "Function URL Setup" above)

## Usage

Once the bot is deployed and the webhook is set up:

1. Start a chat with your bot on Telegram or add it to a group.
2. Send `/start` to get a welcome message.
3. Send `/game` to receive the Snake 4D game.
4. In any chat, type `@your_bot_username` followed by any text to inline query the game.
5. Your highscore will be automatically recorded and displayed in the chat.
6. In group chats, you can compete with other members for the top score.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
