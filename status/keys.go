package status

// TODO: keys:
// TODO:
// TODO: command to check who has keys and is commig today

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"
	"github.com/Schaffenburg/telegram_bot_go/stalk"
	"github.com/Schaffenburg/telegram_bot_go/util"
	"log"
	"strconv"
	"strings"
)

func init() {
	bot := nyu.GetBot()

	bot.Command("ichhabeinenschluessel", handleHaveKey, PermsEV)
	help.AddCommand(tele.Command{
		Text:        "ichhabeinenschluessel",
		Description: "sag bescheid, dass du einen space schluessel hast.",
	})

	bot.Command("ichhabkeinenschluessel", handleDontHaveKey, PermsEV)
	help.AddCommand(tele.Command{
		Text:        "ichhabkeinenschluessel",
		Description: "sag bescheid, dass du keinen space schluessel hast.",
	})

	bot.Command("kommtwermitschluessel", handleListArrivalWKey, perms.MemberSpaceGroup) // no keys
	help.AddCommand(tele.Command{
		Text:        "kommtwermitschluessel",
		Description: "zeigt auf, ob jemand mit schluessel heute kommt.",
	})
}

const TagHasKey = "has_space_key"

func handleHaveKey(m *tele.Message) {
	bot := nyu.GetBot()

	err := db.SetUserTag(m.Sender.ID, TagHasKey)
	if err != nil {
		log.Printf("Failed to set tag %s: %s", TagHasKey, err)
		bot.Sendf(m.Chat, "Ohno, das hat leider nicht funktioniert: %s", err)

		return
	}

	bot.Send(m.Chat, "Ok, merke ich mir!")
}

func handleDontHaveKey(m *tele.Message) {
	bot := nyu.GetBot()

	ch, err := db.RmUserTag(m.Sender.ID, TagHasKey)
	if err != nil {
		log.Printf("Failed to rm tag %s: %s", TagHasKey, err)
		bot.Sendf(m.Chat, "Ohno, das hat leider nicht funktioniert: %s", err)

		return
	}

	if ch {
		bot.Send(m.Chat, "Ok, merke ich mir!")
	} else {
		bot.Send(m.Chat, "Ok, wusste gar nicht, dass du einen hattest o.O")
	}
}

func handleListArrivalWKey(m *tele.Message) {
	bot := nyu.GetBot()

	users, err := ListUsersWithTagArrivingToday(TagHasKey)
	if err != nil {
		log.Printf("Failed to list users with key arriving today: %s", err)
		bot.Sendf(m.Chat, "Ohno, %s", err)

		return
	}

	if len(users) == 0 {
		bot.Send(m.Chat, "Sieht so aus als wuerde noch niemand mit schluessel in den Space kommen.")
	} else {
		b := &strings.Builder{}

		b.WriteString("Ja, es wollen kommen (")
		b.WriteString(strconv.FormatInt(int64(len(users)), 10))
		b.WriteString("):")

		for _, ua := range users {
			b.WriteString("\n - ")

			user, err := stalk.GetUserByID(ua.User)
			if err != nil {
				log.Printf("Failed to query user %d from stalkdb", ua.User)
				b.WriteString(strconv.FormatInt(int64(ua.User), 10))
			}

			b.WriteString(user.FirstName)
			b.WriteString(" ")

			if ua.Arrival.Equal(util.Today(0)) {
				b.WriteString("irgendwann")
			} else {
				b.WriteString(ua.Arrival.Format("15:04"))
			}

		}

		bot.Send(m.Chat, b.String())
	}
}
