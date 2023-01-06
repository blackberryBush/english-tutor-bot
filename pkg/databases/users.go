package databases

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type User struct {
	UserID   int64
	State    int
	LastWord Word
}

func NewUser(userID int64) *User {
	return &User{UserID: userID}
}

func CreateUsers(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users(" +
		"userID	bigint PRIMARY KEY, " +
		"state int, " +
		"lastWord jsonb)")
	if err != nil {
		log.Fatal(err)
	}
}

func ClearUsers(db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		log.Fatal(err)
	}
	CreateUsers(db)
}

func InsertUser(db *sql.DB, user User) {
	c, err := json.Marshal(user.LastWord)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("INSERT INTO users(userID, state, lastWord) VALUES ($1,$2,$3)",
		user.UserID, user.State, c)
	if err != nil {
		log.Println("insert err: ", err)
	}
}

func UpdateUser(db *sql.DB, user User) {
	c, err := json.Marshal(user.LastWord)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = db.Exec("UPDATE users SET state = $1, lastWord = $2 WHERE userID = $3",
		user.State, c, user.UserID)
	if err != nil {
		log.Println("update err: ", err)
	}
}

func GetUser(db *sql.DB, userID int64) (User, error) {
	row := db.QueryRow("SELECT * FROM users WHERE userID = $1", userID)
	if row.Err() != nil {
		return User{}, row.Err()
	}
	user := NewUser(userID)
	var c []byte
	err := row.Scan(&user.UserID, &user.State, &c)
	if err != nil {
		return User{}, errors.New("unknown user")
	}
	err = json.Unmarshal(c, &user.LastWord)
	if err != nil {
		fmt.Println(err)
	}
	return *user, nil
}
