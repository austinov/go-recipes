# tg-bot

Playing with Telegram.

It's a simple telegram-bot, which uses the Telegram API.
It uses *getUpdates* method to retrieve user messages and *sendMessage* method to send replies.
The bot has a few trivial message handlers:
 - **/reverse** - to reverse any text
 - **/search** - to find any text in offered search engine (demonstration of InlineKeyboardMarkup)
 - **/roullete** - to try to guess a number from 1 to 10 (demonstration of ReplyKeyboardMarkup)

Before you run the example you have to register your bot with the BotFather (follow the [manual](https://core.telegram.org/bots)).

To run the bot use:

```
$ go run main.go -t=<token>
```