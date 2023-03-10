package main

import (
	"database/sql"
	bt "english-tutor-bot/pkg/botTutor"
	"english-tutor-bot/pkg/databases"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"os"
)

func getToken() string {
	viper.SetConfigName("token")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		return ""
	}
	return viper.GetString("options.token")
}

func NewPostgresDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=5432 user=postgres dbname=postgres password=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PASSWORD")))
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func main() {
	botAPI, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Fatal(err)
	}
	db, err := NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}
	b := bt.NewTutorBot(botAPI, db)
	databases.CreateUsers(b.DB)
	databases.CreateWords(b.DB)
	//
	//Start timer to send messages&callbacks
	go b.TimeStart()

	// Start checking for updates
	b.Run()
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(b.DB)
}
