package misc

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/stalk"

	"log"
	"strings"
)

const TagGetFood = "get_food"

var (
	LFoodNoone             = loc.MustTrans("misc.food.noone")
	LFoodListGetSth        = loc.MustTrans("misc.food.listgetsth")
	LFoodGetConfirm        = loc.MustTrans("misc.food.get.confirm")
	LFoodGetRevoke         = loc.MustTrans("misc.food.get.revoke")
	LFoodGetRevokeNochange = loc.MustTrans("misc.food.get.revoke.nochange")
)

func init() {
	bot := nyu.GetBot()

	err := db.StartDB()
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err)
	}

	bot.Command("werholtessen", handleWhoGetsFood, PermsDB)
	help.AddCommand("werholtessen")
	bot.Command("ichholeessen", handleIGetFood, PermsDB)
	help.AddCommand("ichholeessen")
	bot.Command("ichholdochkeinessen", handleIDontGetFood, PermsDB)
	help.AddCommand("ichholdochkeinessen")
}

func handleWhoGetsFood(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	u, err := db.GetUsersWithTag(TagGetFood)
	if err != nil {
		log.Printf("Failed to get uses with '%s'-tag: %s", TagGetFood, err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	if len(u) == 0 {
		bot.Send(m.Chat, LFoodNoone.Get(l))

		return
	}

	var user tele.User
	b := strings.Builder{}
	b.WriteString(LFoodListGetSth.Get(l))
	printAnd := false

	for i := 0; i < len(u); i++ {
		user, err = stalk.GetUserByID(u[i])
		if err != nil {
			log.Printf("Failed to get user with id '%d': %s", u[i], err)
			bot.Send(m.Chat, FailGeneric.Getf(l, err))

			return
		}

		if printAnd {
			b.WriteString(", ")
		}

		b.WriteString(user.FirstName)
		if len(user.LastName) > 0 {
			b.WriteString(" ")
			b.WriteString(user.LastName)
		}
		b.WriteString(" (")
		b.WriteString(user.Username)
		b.WriteString(")")

		printAnd = true
	}

	bot.Send(m.Chat, b.String())
}

func handleIGetFood(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetUserTag(m.Sender.ID, TagGetFood)
	if err != nil {
		log.Printf("Failed to set tag '%s' for user %d: %s", TagGetFood, m.Sender.ID, err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	bot.Send(m.Chat, LFoodGetConfirm.Get(l))
}

func handleIDontGetFood(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	ch, err := db.RmUserTag(m.Sender.ID, TagGetFood)
	if err != nil {
		log.Printf("Failed to rm tag '%s' for user %d: %s", TagGetFood, m.Sender.ID, err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	if ch {
		bot.Send(m.Chat, LFoodGetRevoke.Get(l))
	} else {
		bot.Send(m.Chat, LFoodGetRevokeNochange.Get(l))
	}
}
