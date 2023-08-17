/*
this tries to keep an up2date version of all users
*/
package stalk

import (
	db "github.com/Schaffenburg/telegram_bot_go/database"

	tele "gopkg.in/tucnak/telebot.v2"

	"log"
)

func init() {
	database := db.DB()

	_, err := database.Exec("CREATE TABLE IF NOT EXISTS `users` ( `id` BIGINT PRIMARY KEY, `firstname` TEXT, `lastname` TEXT, `username` TEXT, `language` TEXT, `isbot` BOOLEAN );")
	if err != nil {
		log.Fatalf("error creating table users: %s", err)
	}
}

func stalkUsers(m *tele.Message) {
	for _, u := range m.UsersJoined {
		StalkUser(u)
	}

	if m.Sender != nil {
		StalkUser(*m.Sender)
	}
}

func StalkUser(u tele.User) {
	err := UpdateUser(u)
	if err != nil {
		log.Printf("failed to update user while stalking: %s", err)
	}
}

func UpdateUser(u tele.User) error {
	_, err := db.StmtExec(`INSERT INTO users (id, firstname, lastname, username, language, isbot)
	VALUES (?, ?, ?, ?, ?, ?)
	ON DUPLICATE KEY UPDATE
	    firstname = VALUES(firstname),
	    lastname = VALUES(lastname),
	    username = VALUES(username),
	    language = VALUES(language),
	    isbot = VALUES(isbot);`,
		u.ID, u.FirstName, u.LastName, u.Username, u.LanguageCode, u.IsBot,
	)

	return err
}

func GetUserByID(id int64) (t tele.User, err error) {
	r, err := db.StmtQuery("SELECT id, firstname, lastname, username, language, isbot FROM `users` WHERE ( id = ? )",
		id,
	)
	if err != nil {
		return
	}

	if !r.Next() {
		return
	}

	return t, r.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Username, &t.LanguageCode, &t.IsBot)
}

// try ti get user identified by ident
// priority: id > @handle > firstname > lastname
func GetUser(ident string) (t tele.User, ok bool) {
	r, err := db.StmtQuery(`SELECT id, firstname, lastname, username, language, isbot
		FROM users
		WHERE
			id = ?
			OR username = ?
			OR username = SUBSTR(?, 2, LENGTH(?) - 1)
			OR firstname = ?
			OR lastname = ?
		LIMIT 1;`,
		ident, ident, ident, ident, ident, ident,
	)
	if err != nil {
		log.Printf("GetUser(%s) -> %s", ident, err)

		ok = false

		return
	}

	if !r.Next() {
		ok = false

		return
	}

	return t, nil == r.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Username, &t.LanguageCode, &t.IsBot)
}
