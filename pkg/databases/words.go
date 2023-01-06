package databases

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"

	_ "github.com/lib/pq"
	"log"
)

type Word struct {
	UserID  int64
	WordKey int
	TextEN  string
	TextRU  string
}

func NewWord(userID int64, wordKey int, textEN string, textRU string) *Word {
	return &Word{
		UserID:  userID,
		WordKey: wordKey,
		TextEN:  textEN,
		TextRU:  textRU,
	}
}

func CreateWords(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS words(" +
		"userID	bigint, " +
		"wordID int PRIMARY KEY, " +
		"textEN	TEXT, " +
		"textRU	TEXT)")
	if err != nil {
		log.Fatal(err)
	}
}

func ClearWords(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS words")
	if err != nil {
		log.Fatal(err)
	}
	CreateWords(db)
}

func InsertWord(db *sql.DB, word Word) {
	if _, err := GetWord(db, word.TextEN); err == nil {
		return
	}
	_, err := db.Exec("INSERT INTO words(userID, wordID, textEN, textRU) VALUES ($1,$2,$3,$4)",
		word.UserID, word.WordKey, word.TextEN, word.TextRU)
	if err != nil {
		log.Println(err)
	}
}

func GetWord(db *sql.DB, text string) (Word, error) {
	row := db.QueryRow("SELECT * FROM words WHERE textEN = $1 OR textRU = $1", text)
	if row.Err() != nil {
		return Word{}, row.Err()
	}
	word := Word{}
	err := row.Scan(&word.UserID, &word.WordKey, &word.TextEN, &word.TextRU)
	if err != nil {
		return Word{}, errors.New("unknown word")
	}
	return word, err
}

func DeleteWord(db *sql.DB, text string) error {
	_, err := db.Exec("DELETE FROM words WHERE textEN = $1 OR textRU = $1", text)
	if err != nil {
		return err
	}
	return nil
}

func DeleteUser(db *sql.DB, userID int64) error {
	_, err := db.Exec("DELETE FROM words WHERE userID = $1", userID)
	if err != nil {
		return err
	}
	return nil
}

func GetRandomInt(max int) int {
	r, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return -1
	}
	return int(r.Int64())
}

func NewWordID(db *sql.DB) int {
	row := db.QueryRow("SELECT MAX(wordID) FROM words")
	if row.Err() != nil {
		return -1
	}
	id := 0
	err := row.Scan(&id)
	if err != nil {
		return -1
	}
	return id + 1
}

func GetRandomWord(db *sql.DB, userID int64) (Word, error) {
	f := func(rows *sql.Rows) ([]Word, int) {
		var words []Word
		for rows.Next() {
			word := &Word{}
			err := rows.Scan(&word.UserID, &word.WordKey, &word.TextEN, &word.TextRU)
			if err != nil || word.UserID == 0 {
				continue
			}
			words = append(words, *word)
		}
		return words, len(words)
	}
	rows, err := db.Query("SELECT * FROM words WHERE userID = $1 ", userID)
	if err != nil {
		return Word{}, err
	}
	words, n := f(rows)
	if n == 0 {
		return Word{}, errors.New("no data for user")
	}
	return words[GetRandomInt(n)], nil
}

func GetAllWords(db *sql.DB, userID int64) string {
	rows, err := db.Query("SELECT textEN, textRU FROM words WHERE userID = $1", userID)
	if err != nil || rows == nil {
		return ""
	}
	result := "EN\t\t\t| RU\n"
	for rows.Next() {
		var en, ru string
		err := rows.Scan(&en, &ru)
		if err != nil || en == "" {
			continue
		}
		result += fmt.Sprintf("%v\t\t\t| %v\n", en, ru)
	}
	return result
}
