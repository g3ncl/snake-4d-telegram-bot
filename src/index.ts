import { APIGatewayProxyEvent, APIGatewayProxyHandler } from "aws-lambda";
import TelegramBot from "node-telegram-bot-api";
import winston from "winston";

// Enable logging
const logger = winston.createLogger({
  level: "info",
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.printf(({ timestamp, level, message }) => {
      return `${timestamp} - ${level}: ${message}`;
    })
  ),
  transports: [new winston.transports.Console()],
});

const GAME_SHORT_NAME: string = "snake4d";
const GAME_URL: string = "https://snake4d.netlify.app";
const TOKEN = process.env.TELEGRAM_BOT_TOKEN;

if (!TOKEN) {
  logger.error("TELEGRAM_BOT_TOKEN is not set in environment variables");
  throw new Error("TELEGRAM_BOT_TOKEN is not set");
}

const bot = new TelegramBot(TOKEN);

export const webhook: APIGatewayProxyHandler = async (
  event: APIGatewayProxyEvent
) => {
  if (!event.body) {
    return { statusCode: 400, body: "Invalid request body" };
  }

  const update = JSON.parse(event.body);

  if (update.message && update.message.text) {
    const chatId = update.message.chat.id;
    const messageText = update.message.text;

    if (messageText === "/start") {
      await bot.sendMessage(
        chatId,
        "Welcome to the Snake 4D Game Bot! Use @snake4dbot followed by some text in any chat to start playing."
      );
    } else if (messageText === "/game") {
      await bot.sendGame(chatId, GAME_SHORT_NAME);
    }
  } else if (update.inline_query) {
    const results: TelegramBot.InlineQueryResultGame[] = [
      {
        type: "game",
        id: GAME_SHORT_NAME,
        game_short_name: GAME_SHORT_NAME,
      },
    ];
    await bot.answerInlineQuery(update.inline_query.id, results);
  } else if (
    update.callback_query &&
    update.callback_query.game_short_name === GAME_SHORT_NAME
  ) {
    await bot.answerCallbackQuery(update.callback_query.id, {
      url: `${GAME_URL}/#userId=${update.callback_query.from.id}&messageId=${update.callback_query.inline_message_id}`,
    });
  }

  return { statusCode: 200, body: "OK" };
};

logger.info("Lambda function is ready to process events...");
