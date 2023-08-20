/*
this tries to keep an up2date version of all groupMembers
*/
package stalk

import (
	tele "gopkg.in/tucnak/telebot.v2"

	//"github.com/Schaffenburg/telegram_bot_go/nyu"

	db "github.com/Schaffenburg/telegram_bot_go/database"

	"log"
)

// returns if user is member of group
func IsMember(user, group int64) (bool, error) {
	r, err := db.StmtQuery("SELECT * FROM memberships WHERE user = ? AND group_id = ?;",
		user, group,
	)
	if err != nil {
		return false, err
	}

	return r.Next(), nil
}

// returns if user is member of group
func IsTaggedGroupMember(user int64, tag string) (bool, error) {
	r, err := db.StmtQuery("SELECT 1 FROM memberships WHERE group_id = (SELECT group_id FROM group_tags WHERE tag = ?) AND user = ?",
		tag, user,
	)
	if err != nil {
		return false, err
	}

	var count int64
	if !r.Next() {
		return false, nil
	}

	err = r.Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func stalkMemberships(m *tele.Message) {
	memberships := make([]int64, len(m.UsersJoined))

	if m.UsersJoined != nil {
		for i, u := range m.UsersJoined {
			memberships[i] = u.ID
		}
	}

	if len(memberships) > 0 {
		err := UpdateMembership(m.Chat.ID, memberships...)
		if err != nil {
			log.Printf("Error updating memberships of %v for chat %d: %s", memberships, m.Chat.ID, err)
		}
	}

	if m.UserLeft != nil {
		ok, err := UpdateMembershipDelete(m.Chat.ID, m.UserLeft.ID)
		if err != nil {
			log.Printf("Error updating memberships of %v for chat %d: %s", memberships, m.Chat.ID, err)
		}

		if !ok {
			log.Printf("User [%d] left [%d] without me noticing them joining o.O", m.UserLeft.ID, m.Chat.ID)
		}
	}

	if len(memberships) > 0 {
		err := UpdateMembership(m.Chat.ID, memberships...)
		if err != nil {
			log.Printf("Error updating memberships of %v for chat %d: %s", memberships, m.Chat.ID, err)
		}
	}

	if m.Chat != nil && m.Sender != nil && m.Chat.Type == tele.ChatGroup {
		UpdateMembership(m.Chat.ID, m.Sender.ID)
	}
}

func UpdateMembershipDelete(group, user int64) (ch bool, err error) {
	r, err := db.StmtExec("DELETE FROM memberships WHERE group_id = ? AND user = ?;",
		group, user,
	)
	if err != nil {
		return
	}

	i, err := r.RowsAffected()

	return i > 0, err
}

func UpdateMembership(group int64, users ...int64) error {
	for _, u := range users {
		_, err := db.StmtExec(`INSERT INTO memberships (group_id, user)
			VALUES (?, ?)
			ON DUPLICATE KEY UPDATE
			    group_id = VALUES(group_id),
			    user = VALUES(user);`,
			group, u,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func GetUserIDsByGroupID(id int64) (users []int64, err error) {
	r, err := db.StmtQuery("SELECT user FROM memberships WHERE group_id = ?", id)
	if err != nil {
		return nil, err
	}

	var buf int64
	users = make([]int64, 0)

	for r.Next() {
		if err = r.Scan(&buf); err != nil {
			return nil, err
		} else {
			users = append(users, buf)
		}
	}

	return users, nil
}

/*func GetUsersByGroupID(id int64) (t []tele.User) {
	r, err := db.StmtQuery("SELECT id, firstname, lastname, username, language, isbot FROM `users` WHERE ( id = ? )",
		id,
	)

	if !r.Next() {
		return
	}

	return t, r.Scan(&t.ID, &t.FirstName, &t.LastName, &t.Username, &t.LanguageCode, &t.IsBot)
}
*/
