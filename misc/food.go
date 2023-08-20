package misc

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/stalk"

	"log"
	"strings"
)

func init() {
	bot := nyu.GetBot()

	err := db.StartDB()
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err)
	}

	database := db.DB()

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `foods` ( `user` BIGINT, `food` VARCHAR(50), `preferance` TEXT, PRIMARY KEY (user, food));")
	if err != nil {
		log.Printf("error creating table foods: %s", err)
		return
	}

	// Set
	bot.Command("/ichmag", handleSetFavFood, PermsDB)
	bot.Command("/ichmagnicht", handleRmFavFood, PermsDB)

	// query
	bot.Command("/wasmag", handleWhatLikes, PermsDB)
	bot.Command("/wasmagich", handleWhatDoILike, PermsDB)
}

func handleSetFavFood(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.Split(m.Text, " ")
	if len(args) < 3 {
		bot.Send(m.Chat, "Usage: /ichmag <food> <desc...>")
		return
	}

	food := args[1]
	desc := strings.Join(args[2:], " ")

	err := SetFavFood(m.Sender.ID, food, desc)
	if err != nil {
		log.Printf("failed to set fav food for %d: %s", m.Sender.ID, err)
		bot.Send(m.Chat, "Ohno, es gab einen Fehler: "+err.Error())
	} else {
		bot.Send(m.Chat, "Noted.")
	}
}

func handleRmFavFood(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.Split(m.Text, " ")
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /ichmagnicht <food>")
		return
	}

	food := args[1]

	err, changed := RmFavFood(m.Sender.ID, food)
	if err != nil {
		log.Printf("failed to remove fav food for %d: %s", m.Sender.ID, err)
		bot.Send(m.Chat, "Ohno, es gab einen Fehler: "+err.Error())
	} else {
		if changed {
			bot.Send(m.Chat, "Noted.")
		} else {
			bot.Send(m.Chat, "Wusste ich gar, dass du das mochtest o.O.")
		}
	}
}

type Food struct {
	Food, Preferance string
}

func GetFavFoods(u int64) ([]Food, error) {
	d, err := db.StmtQuery("SELECT food, preferance FROM foods WHERE user = ?", u)
	if err != nil {
		return nil, err
	}

	f := make([]Food, 0)
	var food, pref string

	for d.Next() {
		err := d.Scan(&food, &pref)
		if err != nil {
			return nil, err
		}

		f = append(f, Food{food, pref})
	}

	return f, nil
}

func SetFavFood(u int64, food, desc string) error {
	_, err := db.StmtExec("INSERT INTO foods (user, food, preferance) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE preferance = VALUES(preferance);",
		u, food, desc,
	)

	return err
}

func RmFavFood(u int64, food string) (err error, changed bool) {
	r, err := db.StmtExec("DELETE FROM foods WHERE (user = ?) AND (food = ?);",
		u, food,
	)

	if err != nil {
		return
	}

	num, err := r.RowsAffected()

	return err, num > 0
}

func handleWhatLikes(m *tele.Message) {
	const usage = "Usage: /wasmag @mention [food]"

	bot := nyu.GetBot()

	args := strings.Split(m.Text, " ")
	if len(args) < 2 {
		bot.Send(m.Chat, usage)
		return
	}

	user, ok := stalk.GetUser(args[1])
	if !ok {
		bot.Send(m.Chat, ""+args[1]+" konnte nicht gefunden werden :(")
		return
	}

	f, err := GetFavFoods(user.ID)
	if err != nil {
		log.Printf("Failed to get Likes of %d: %s", user.ID, err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
		return
	}

	if len(f) <= 0 {
		bot.Send(m.Chat, user.FirstName+" hat (noch) keine Vorlieben eingetragen.")
		return
	}

	b := &strings.Builder{}

	b.WriteString(user.FirstName)
	b.WriteString(" (")
	b.WriteString(user.Username)
	b.WriteString(") mag")

	if len(args) == 3 {
		fnd := false
		b.WriteString(" ")
		b.WriteString(args[2])

		for _, f := range f {
			if f.Food == args[2] {
				fnd = true

				b.WriteString(" (")
				b.WriteString(f.Preferance)
				b.WriteString(")")
			}
		}

		if !fnd {
			b.WriteString(" nicht, das ist mobbing!")
		}
	} else {
		b.WriteString(":")

		for _, f := range f {
			b.WriteString("\n - ")
			b.WriteString(f.Food)
			b.WriteString(" (")
			b.WriteString(f.Preferance)
			b.WriteString(")")
		}
	}

	bot.Send(m.Chat, b.String())
}

func handleWhatDoILike(m *tele.Message) {
	bot := nyu.GetBot()

	f, err := GetFavFoods(m.Sender.ID)
	if err != nil {
		log.Printf("Failed to get Likes of %d: %s", m.Sender.ID, err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
		return
	}

	if len(f) <= 0 {
		bot.Send(m.Chat, "Du mochtest doch gar nichts o.O")
		return
	}

	b := &strings.Builder{}

	b.WriteString("Do magst:")

	for _, f := range f {
		b.WriteString("\n - ")
		b.WriteString(f.Food)
		b.WriteString(" (")
		b.WriteString(f.Preferance)
		b.WriteString(")")
	}

	bot.Send(m.Chat, b.String())
}
