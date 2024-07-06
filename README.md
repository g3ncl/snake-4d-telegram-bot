# Snake 4D Telegram Bot

This repository contains the code for a Telegram bot that allows users to play the Snake 4D game directly within Telegram chats. The bot is built using Node.js and is designed to be deployed on AWS Lambda.

## Features

- Responds to `/start` command with a welcome message
- Responds to `/game` command by sending the Snake 4D game
- Supports inline queries to start the game in any chat
- Integrates with the Snake 4D game hosted at https://snake4d.netlify.app
- Utilizes Telegram's Game API to set and display highscores within chats

## Prerequisites

- Node.js (v20 or later)
- npm (Node Package Manager)
- AWS account with Lambda and IAM access
- Telegram Bot Token

## Setup

1. Clone this repository:

   ```
   git clone https://github.com/g3ncl/snake-4d-telegram-bot.git
   cd snake-4d-telegram-bot
   ```

2. Install dependencies:

   ```
   npm install
   ```

3. Create a `.env` file in the root directory and add your Telegram Bot Token:

   ```
   TELEGRAM_BOT_TOKEN= your_bot_token
   ```

4. Build the TypeScript code:
   ```
   npm run build
   ```

## Deployment

This bot is configured to be automatically deployed to AWS Lambda using GitHub Actions. The workflow is triggered on every push to the `main` branch.

To enable the automatic deployment, you need to set up the following secrets in your GitHub repository:

- `AWS_ACCESS_KEY_ID`: Your AWS access key ID
- `AWS_SECRET_ACCESS_KEY`: Your AWS secret access key
- `AWS_REGION`: The AWS region where your Lambda function is located
- `TELEGRAM_BOT_TOKEN`: Your Telegram Bot Token
- `LAMBDA_EXECUTION_ROLE`: The ARN of the IAM role for Lambda execution

### Manual Deployment

If you prefer to deploy manually or need to set up the Lambda function for the first time:

1. Zip the contents of the `dist` folder and `node_modules`:

   ```
   zip -r function.zip dist node_modules
   ```

2. Use the AWS CLI or AWS Management Console to create or update your Lambda function with the `function.zip` file.

3. Set the environment variable `TELEGRAM_BOT_TOKEN` in your Lambda function configuration.

4. Configure the Lambda function handler as `dist/index.webhook`.

5. Set up an API Gateway trigger for your Lambda function and configure the Telegram webhook to point to the API Gateway URL.

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
