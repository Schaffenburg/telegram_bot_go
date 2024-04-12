package status

// TODO: keys:
// TODO:
// TODO: command to check who has keys and is commig today

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
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
	help.AddCommand("ichhabeinenschluessel")

	bot.Command("ichhabkeinenschluessel", handleDontHaveKey, PermsEV)
	help.AddCommand("ichhabkeinenschluessel")

	bot.Command("kommtwermitschluessel", handleListArrivalWKey, perms.MemberSpaceGroup) // no keys
	help.AddCommand("kommtwermitschluessel")
}

var (
	LKeyHaveConfirm             = loc.MustTrans("status.key.have.confirm")
	LKeyDontHaveConfirm         = loc.MustTrans("status.key.donthave.confirm")
	LKeyDontHaveConfirmNochange = loc.MustTrans("status.key.donthave.confirm.nochange")
	LKeyStatusNoone             = loc.MustTrans("status.key.status.noone")
	LKeyStatusList              = loc.MustTrans("status.key.status.list")
	LKeyStatusSometime          = loc.MustTrans("status.key.status.sometime")
)

const TagHasKey = "has_space_key"

func handleHaveKey(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetUserTag(m.Sender.ID, TagHasKey)
	if err != nil {
		log.Printf("Failed to set tag %s: %s", TagHasKey, err)
		bot.Sendf(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	bot.Send(m.Chat, LKeyHaveConfirm.Get(l))
}

func handleDontHaveKey(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	ch, err := db.RmUserTag(m.Sender.ID, TagHasKey)
	if err != nil {
		log.Printf("Failed to rm tag %s: %s", TagHasKey, err)
		bot.Sendf(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	if ch {
		bot.Send(m.Chat, LKeyDontHaveConfirm.Get(l))
	} else {
		bot.Send(m.Chat, LKeyDontHaveConfirmNochange.Get(l))
	}
}

func handleListArrivalWKey(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	users, err := ListUsersWithTagArrivingToday(TagHasKey)
	if err != nil {
		log.Printf("Failed to list users with key arriving today: %s", err)
		bot.Sendf(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	if len(users) == 0 {
		bot.Send(m.Chat, LKeyStatusNoone.Get(l))
	} else {
		b := &strings.Builder{}

		b.WriteString(LKeyStatusList.Getf(l, len(users)))

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
				b.WriteString(LKeyStatusSometime.Get(l))
			} else {
				b.WriteString(ua.Arrival.Format("15:04"))
			}

		}

		bot.Send(m.Chat, b.String())
	}
}
