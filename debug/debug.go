package debug

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
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

	// permissions required for debug commands:
	const permFailText = "Debugging Befehle benoetigen sowohl das debug Tag als auch eine Mitgliedschaft in der e.V. gruppe"
	perms := []nyu.Permission{
		&nyu.PermissionFailText{Perm: &perms.PermissionTag{"debug"}, Text: permFailText},
		&nyu.PermissionFailText{Perm: &perms.PermissionGroupTag{"perm_ev"}, Text: permFailText},
	}

	bot.Command("debug_settagself", handleSetTagSelf, perms...)
	bot.Command("debug_rmtagself", handleRmTagSelf, perms...)

	bot.Command("debug_setgrouptagcurrent", handleSetGroupTagCurrent, perms...)

	bot.Command("debug_isgroupmemberself", handleIsGroupMemberSelf, perms...)
	bot.Command("debug_istaggedgroupmemberself", handleIsTaggedGroupMemberSelf, perms...)

	bot.AnswerCommand("hi", "Hi to you too, %u!")
	bot.ReplyCommand("nya", "nya")

	poller := nyu.Poller()

	ch := make(chan tele.Update, 100)
	poller.AddCh(ch)

	go func() {
		var u tele.Update
		var ok bool

		for {
			u, ok = <-ch
			if !ok {
				return
			}

			if u.Message != nil {
				nyu.LogMessage("rec <-", u.Message)
			}
		}
	}()

	// check for any unhandled commands
	bot.Handle(tele.OnText, func(m *tele.Message) {
		if len(m.Text) > 0 && m.Text[0] == '/' {
			bot.Send(m.Chat, "Invalid Command!")
		}
	})
}

func handleRmTagSelf(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Sender, "Usage: /debug_rmtagself <tag>")

		return
	}

	tag := args[1]

	changed, err := db.RmUserTag(m.Sender.ID, tag)
	if err != nil {
		log.Printf("Error removing user tag: %s", err)
		bot.Send(m.Sender, "Ohno, "+err.Error())
	} else {
		if changed {
			bot.Send(m.Sender, "Removed tag.")
		} else {
			bot.Send(m.Sender, "Tag was not present.")
		}
	}
}

func handleSetTagSelf(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_settagself <tag>")

		return
	}

	tag := args[1]

	err := db.SetUserTag(m.Sender.ID, tag)
	if err != nil {
		log.Printf("Error adding user tag: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "Set tag.")
	}
}

func handleIsGroupMemberSelf(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_isgroupmemberself <group(int64)>")

		return
	}

	group, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		bot.Send(m.Chat, "Ohno, "+err.Error())
	}

	status, err := stalk.IsMember(m.Sender.ID, group)
	if err != nil {
		log.Printf("Error checking membership: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "You are"+util.T(status, "", "n't")+" a member.")
	}
}

func handleSetGroupTagCurrent(m *tele.Message) {
	bot := nyu.GetBot()

	if m.Chat.Type != tele.ChatGroup {
		bot.Send(m.Chat, "grouptags are only available for Chat.Type \"ChatGroup\".")
		return
	}

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_setgrouptagcurrent <tag[len:20]>")

		return
	}

	tag := args[1]

	err := db.SetGroupTag(m.Chat.ID, tag)
	if err != nil {
		log.Printf("Error setting group tag: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "Set tag.")
	}
}

func handleIsTaggedGroupMemberSelf(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_istaggedgroupmemberself <tag[len:20]>")

		return
	}

	tag := args[1]

	status, err := stalk.IsTaggedGroupMember(m.Sender.ID, tag)
	if err != nil {
		log.Printf("Error getting membership status of tagged group: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "You are"+util.T(status, "", "n't")+" a member.")
	}
}
