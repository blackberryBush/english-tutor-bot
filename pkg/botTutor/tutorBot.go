package botTutor

import (
	"database/sql"
	bb "english-tutor-bot/pkg/botBasic"
	yt "github.com/blackberryBush/yandex-translate"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/spf13/viper"
	"log"
	"time"
)

type TutuorBot struct {
	bb.BotGeneral
	Chapters    []int
	iterations  int
	timers      map[int64]*time.Timer
	yTranslator yt.YandexTranslator
}

func NewTutorBot(bot *tgbotapi.BotAPI, dbTasks *sql.DB) *TutuorBot {
	var _ bb.BotInterface = &TutuorBot{}
	viper.SetConfigName("token")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err == nil {
		return &TutuorBot{
			BotGeneral: *bb.NewBot(bot, dbTasks),
			Chapters:   nil,
			timers:     make(map[int64]*time.Timer),
			yTranslator: *yt.NewYandexTranslator(viper.GetString("options.folderID"),
				viper.GetString("options.oauthKey"), 10*time.Minute),
		}
	}
	return nil
}

func (b *TutuorBot) Run() {
	log.Printf("Authorized on account %s", b.Bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.Bot.GetUpdatesChan(u)
	b.Bot.Self.CanJoinGroups = false
	b.SendCommands(tgbotapi.BotCommand{
		Command:     "/repeat",
		Description: "Напоминалка случайных слов",
	}, tgbotapi.BotCommand{
		Command:     "/quiz",
		Description: "Викторина",
	}, tgbotapi.BotCommand{
		Command:     "/clearme",
		Description: "Удалить все слова для себя",
	}, tgbotapi.BotCommand{
		Command:     "/clearall",
		Description: "Удалить все слова для всех",
	}, tgbotapi.BotCommand{
		Command:     "/delword",
		Description: "Удалить слово (en)",
	}, tgbotapi.BotCommand{
		Command:     "/addword",
		Description: "Добавить слово: en - ru",
	}, tgbotapi.BotCommand{
		Command:     "/getallwords",
		Description: "Получить список всех слов",
	})
	for update := range updates {
		b.HandleUpdate(&update)
	}
}

func lastSend(chatID int64, messageTimes map[int64]time.Time) bool {
	if val, ok := messageTimes[chatID]; ok {
		dt := time.Now()
		return dt.After(val.Add(time.Second))
	}
	return true
}

func (b *TutuorBot) TimeStart() {
	messageTimes := make(map[int64]time.Time)
	timer := time.NewTicker(time.Second / 30)
	defer timer.Stop()
	for range timer.C {
		b.SendQueue.Range(func(i int64, v bb.ItemToSend) bool {
			if v.Queue > 0 && lastSend(i, messageTimes) {
				err := b.Send(i)
				if err != nil {
					log.Println(err)
				}
				messageTimes[i] = time.Now()
				return false
			}
			if v.Queue <= 0 {
				b.SendQueue.Delete(i)
			}
			return true
		})
	}
}
