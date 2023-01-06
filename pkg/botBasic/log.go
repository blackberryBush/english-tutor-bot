package botBasic

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"strings"
)

const (
	messageUndefined = -1
	messageCommand   = iota
	messageText
)

func PrintReceive(update *tgbotapi.Update, updateType int, chatID int64) {
	user := update.SentFrom()
	ans := "[idk who is this]"
	if user != nil {
		ans = fmt.Sprintf("[@%s][%s", user.UserName, user.FirstName)
		if user.LastName != "" {
			ans = fmt.Sprintf("%s %s", ans, user.LastName)
		}
		ans = fmt.Sprintf("%s][%d]", ans, chatID)
	}
	switch updateType {
	case messageCommand:
		log.Printf("NEW COMMAND:	%s %s\n", ans, update.Message.Text)
	case messageText:
		log.Printf("NEW MESSAGE:	%s %s\n", ans, update.Message.Text)
	default:
		log.Printf("NEW INDEFINITE:	%s\n", ans)
	}
}

func PrintSent(c *tgbotapi.Chattable) {
	switch (*c).(type) {
	case tgbotapi.MessageConfig:
		text := strings.Replace((*c).(tgbotapi.MessageConfig).Text, "\n", " -> ", -1)
		log.Printf("SENT MESSAGE:	[%d] %s", (*c).(tgbotapi.MessageConfig).ChatID, text)
	case tgbotapi.CallbackConfig:
		log.Printf("SENT CALLBACK:	[%s] %s", (*c).(tgbotapi.CallbackConfig).CallbackQueryID, (*c).(tgbotapi.CallbackConfig).Text)
	case tgbotapi.DeleteMessageConfig:
		log.Printf("DELETE MESSAGE:	[%d] %d", (*c).(tgbotapi.DeleteMessageConfig).ChatID, (*c).(tgbotapi.DeleteMessageConfig).MessageID)
	default:
		log.Println("SENT, but idk what is this")
	}
}
