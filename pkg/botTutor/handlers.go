package botTutor

import (
	bb "english-tutor-bot/pkg/botBasic"
	"english-tutor-bot/pkg/databases"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

// ТЗ: таблица юзерid - слово en - слово ru +
// добавить/удалить слово +
// удалить все слова для cебя +>(by userID)
// удалить вообще если ты админ + (need adding check if u'r admin)

const (
	messageUndefined = -1
	messageCommand   = iota
	messageText
)

func (b *TutuorBot) getUpdateType(update *tgbotapi.Update) (int, int64, databases.User, error) {
	if update.Message != nil {
		chatID := update.Message.From.ID
		user, err := databases.GetUser(b.DB, chatID)
		if update.Message.IsCommand() {
			return messageCommand, chatID, user, err
		}
		if update.Message.Text != "" {
			return messageText, chatID, user, err
		}
		return messageUndefined, chatID, user, err
	}
	return messageUndefined, 0, databases.User{}, nil
}

func (b *TutuorBot) HandleUpdate(update *tgbotapi.Update) {
	updateType, chatID, currentUser, err := b.getUpdateType(update)
	if err != nil {
		currentUser = *databases.NewUser(chatID)
		databases.InsertUser(b.DB, currentUser)
	}
	bb.PrintReceive(update, updateType, chatID)
	switch updateType {
	case messageCommand:
		b.handleCommand(update.Message, &currentUser)
	case messageText:
		b.handleText(update.Message, &currentUser)
	default:
		b.handleUnknown(&currentUser)
	}
	databases.UpdateUser(b.DB, currentUser)
}

func compareWithAdmins(id int64) bool {
	viper.SetConfigName("token")
	viper.AddConfigPath(".")
	var admins []string
	if err := viper.ReadInConfig(); err == nil {
		admins = viper.GetStringSlice("options.admins")
	}
	for _, v := range admins {
		if strconv.FormatInt(id, 10) == v {
			return true
		}
	}
	return false
}

func (b *TutuorBot) handleCommand(message *tgbotapi.Message, user *databases.User) {
	chatID := user.UserID
	switch message.CommandWithAt() {
	case "start":
		b.PullText("Основные команды бота:\n", chatID, 0)
	case "clearme":
		databases.DeleteUser(b.DB, chatID)
		b.PullText("Очищено.", chatID, 0)
	case "clearall":
		if compareWithAdmins(user.UserID) {
			databases.ClearWords(b.DB)
			databases.ClearUsers(b.DB)
			b.PullText("Успешно.", chatID, 0)
		} else {
			b.PullText("Вы не админ! (Ваш ID "+strconv.FormatInt(chatID, 10)+")", chatID, message.MessageID)
		}
	case "delword":
		if message.CommandArguments() == "" {
			b.PullText("Не введено слово", chatID, 0)
			return
		}
		databases.DeleteWord(b.DB, message.CommandArguments())
		b.PullText("Очищено.", chatID, 0)
	case "addword":
		arr := strings.Split(message.CommandArguments(), " ")
		if len(arr) < 2 {
			b.PullText("Не введено слово", chatID, 0)
			return
		}
		word := databases.NewWord(chatID, databases.NewWordID(b.DB), arr[0], arr[1])
		databases.InsertWord(b.DB, *word)
		b.PullText("Добавлено.", chatID, 0)
	case "repeat":
		currentWord, err := databases.GetRandomWord(b.DB, user.UserID)
		if err != nil {
			b.PullText("Произошла ошибка получения слова.", chatID, 0)
		} else {
			b.PullText(currentWord.TextEN+" = "+currentWord.TextRU, chatID, 0)
		}
	case "quiz":
		user.State = 2
		b.newQuizWord(user, "Для остановки викторины напишите слово stop\n\n")
	case "getallwords":
		s := databases.GetAllWords(b.DB, user.UserID)
		b.PullFileBytes([]byte(s), chatID, message.MessageID, "notes.txt")
	default:
		b.handleUnknown(user)
	}
}

func getLanguage(text string) string {
	for _, v := range text {
		if v >= 'A' && v <= 'Z' || v >= 'a' && v <= 'z' {
			return "en"
		} else if v >= 'а' && v <= 'я' || v >= 'А' && v <= 'Я' {
			return "ru"
		}
	}
	return ""
}

func (b *TutuorBot) handleText(message *tgbotapi.Message, user *databases.User) {
	switch user.State {
	case 0:
		b.handleTranslate(message, user)
	case 1:
		b.handleAdding(message, user)
	case 2:
		b.handleQuiz(message, user)
	}
}

func (b *TutuorBot) newQuizWord(user *databases.User, msg string) {
	currentWord, err := databases.GetRandomWord(b.DB, user.UserID)
	if err != nil {
		b.PullText("Произошла ошибка получения слова.", user.UserID, 0)
		user.State = 0
	} else {
		b.PullText(msg+"Переведите на английский: "+currentWord.TextRU, user.UserID, 0)
		user.LastWord = currentWord
	}
}

func (b *TutuorBot) handleQuiz(message *tgbotapi.Message, user *databases.User) {
	if strings.EqualFold(message.Text, "stop") {
		user.State = 0
		b.PullText("Викторина остановлена.", user.UserID, 0)
		return
	}
	if strings.EqualFold(message.Text, user.LastWord.TextEN) {
		b.newQuizWord(user, "Верно.\n\n")
	} else {
		b.newQuizWord(user, "Неверно.\n\n")
	}
}

func (b *TutuorBot) handleTranslate(message *tgbotapi.Message, user *databases.User) {
	chatID := user.UserID
	switch getLanguage(message.Text) {
	case "en":
		tr, err := b.yTranslator.TranslateByYandex("ru", message.Text)
		if err != nil {
			b.PullText(fmt.Sprintf("Ошибка перевода: %s", err), chatID, message.MessageID)
		} else {
			user.LastWord = *databases.NewWord(chatID, databases.NewWordID(b.DB), message.Text, tr)
			keyboard := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(`Да`),
				tgbotapi.NewKeyboardButton(`Нет`))
			user.State = 1
			b.PullText("Перевод: "+tr+"\n\nДобавить в словарь?", chatID, message.MessageID, keyboard)
		}
	case "ru":
		tr, err := b.yTranslator.TranslateByYandex("en", message.Text)
		if err != nil {
			b.PullText(fmt.Sprintf("Ошибка перевода: %s", err), chatID, message.MessageID)
		} else {
			user.LastWord = *databases.NewWord(chatID, databases.NewWordID(b.DB), tr, message.Text)
			keyboard := tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton(`Да`),
				tgbotapi.NewKeyboardButton(`Нет`))
			user.State = 1
			b.PullText("Перевод: "+tr+"\n\nДобавить в словарь?", chatID, message.MessageID, keyboard)
		}
	default:
		b.PullText("Ошибка определения языка.", chatID, message.MessageID)
	}
}

func (b *TutuorBot) handleAdding(message *tgbotapi.Message, user *databases.User) {
	chatID := user.UserID
	if user.State == 1 {
		switch message.Text {
		case "Да", "да", "Да.":
			databases.InsertWord(b.DB, user.LastWord)
			user.State = 0
			b.PullText("Успешно добавлено в словарь.", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
		case "Нет", "нет", "Нет.":
			user.State = 0
			b.PullText("Слово не будет добавлено в словарь.", chatID, 0, tgbotapi.ReplyKeyboardRemove{RemoveKeyboard: true})
		}
		return
	}
}

func (b *TutuorBot) handleUnknown(user *databases.User) {
	b.PullText("Я не знаю, что это))", user.UserID, 0)
}
