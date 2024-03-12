package debug

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	loc "github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"
	"github.com/Schaffenburg/telegram_bot_go/stalk"
	"github.com/Schaffenburg/telegram_bot_go/status"
	"github.com/Schaffenburg/telegram_bot_go/util"

	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func init() {
	bot := nyu.GetBot()

	// permissions required for debug commands:
	permFailText := loc.MustTrans("parms.FailDebug")
	perms := []nyu.Permission{
		&nyu.PermissionFailText{Perm: &perms.PermissionTag{"debug"}, Text: permFailText},
		&nyu.PermissionFailText{Perm: &perms.PermissionGroupTag{"perm_ev"}, Text: permFailText},
	}

	bot.Command("debug_settagself", handleSetTagSelf, perms...)
	bot.Command("debug_rmtagself", handleRmTagSelf, perms...)

	bot.Command("debug_setgrouptagcurrent", handleSetGroupTagCurrent, perms...)

	bot.Command("debug_isgroupmemberself", handleIsGroupMemberSelf, perms...)
	bot.Command("debug_istaggedgroupmemberself", handleIsTaggedGroupMemberSelf, perms...)

	bot.Command("debug_testinlinebtn", handleTestInline, perms...)

	bot.Command("debug_teststatusinline", handleTestStatusInline, perms...)

	bot.Command("debug_leave", handleLeave, perms...)

	bot.Command("debug_importsubscriptions", handleImportSubscriptions, perms...)
	bot.Command("debug_importlangmap", handleImportLanguageMap, perms...)

	debugcallback := func(c *tele.Callback) {
		nyu.GetBot().Reply(c.Message, strconv.Quote(c.Data))
	}

	bot.HandleInlineCallback("testinline_ok", debugcallback)
	bot.HandleInlineCallback("testinline_nu", debugcallback)

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
				nyu.LogMessage("msg <-", u.Message)
			}

			if u.Callback != nil {
				nyu.LogCallback("cb <-", u.Callback)
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

func handleImportLanguageMap(m *tele.Message) {
	bot := nyu.GetBot()

	streamer, err := bot.NewEditStreamer(m.Chat, "create editstreamer")
	if err != nil {
		bot.Send(m.Chat, "Ohno, failed to create editstreamer: %s", err)
		log.Printf("Failed to create editstreamer %s", err)

		return
	}

	log := func(f string, a ...any) {
		str := fmt.Sprintf(f, a...) + "\n"

		log.Printf(str)
		streamer.Append(str)
	}

	log("Starting language import from `languagemap`")

	log("NO!! language is not mapped yet, pls do that first!")
	return

	// TODO: map language
	langmap := map[int]string{ // or vice versa idunno
		0: "Deutsch",
		1: "English",
	}

	s, err := db.StmtQuery("SELECT (user_id, language) FROM languagemap")
	if err != nil {
		log("failed to get spacestatus from DB: %s", err)
	}

	defer s.Close()

	var user int64
	var language int

	for s.Next() {
		err = s.Scan(&user, &language)
		if err != nil {
			log("Failed to scan: %s", err)

			return
		}

		err := loc.SetUserLanguage(user,
			*loc.GetLanguage(langmap[language]),
		) // TODO: index lanugage
		if err != nil {
			log("Failed to insert: %s", err)

			return
		}

	}
}

// TODO: test RENAME old spacestatus table to spacestatus_old!!!!!!
/*
	CREATE TABLE `spacestatus` (
	  `status_id` int(11) NOT NULL AUTO_INCREMENT,
	  `status_short` tinyint(1) NOT NULL,
	  `status_full` int(11) NOT NULL,
	  `status_timestamp` timestamp NOT NULL DEFAULT current_timestamp(),
	  PRIMARY KEY (`status_id`)
	) ENGINE=InnoDB AUTO_INCREMENT=158719 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci ROW_FORMAT=COMPACT;
*/
func handleImportStatus(m *tele.Message) {
	bot := nyu.GetBot()

	streamer, err := bot.NewEditStreamer(m.Chat, "create editstreamer")
	if err != nil {
		bot.Send(m.Chat, "Ohno, failed to create editstreamer: %s", err)
		log.Printf("Failed to create editstreamer %s", err)

		return
	}

	log := func(f string, a ...any) {
		str := fmt.Sprintf(f, a...) + "\n"

		log.Printf(str)
		streamer.Append(str)
	}

	log("Starting spacestatus import from `spacestatus_old`")

	s, err := db.StmtQuery("SELECT (status_full, status_timestamp) FROM spacestatus_old ORDER BY status_timestamp ASC")
	if err != nil {
		log("failed to get spacestatus from DB: %s", err)
	}

	defer s.Close()

	var status string
	var timestamp time.Time

	for s.Next() {
		err = s.Scan(&status, &timestamp)
		if err != nil {
			log("Failed to scan: %s", err)

			return
		}

		_, err := db.StmtExec("INSERT INTO spacestatus (time, status) VALUES (?, ?);",
			timestamp.Unix(), status,
		)

		if err != nil {
			log("Failed to insert: %s", err)

			return
		}

	}

	log("Döner!")
}

// TODO: test
/*
	CREATE TABLE `subscription` (
	  `subscription_id` int(11) NOT NULL AUTO_INCREMENT,
	  `user_id` bigint(20) DEFAULT NULL,
	  `type` int(1) DEFAULT NULL,
	  `updated` datetime DEFAULT NULL,
	  PRIMARY KEY (`subscription_id`)
	);
*/
func handleImportSubscriptions(m *tele.Message) {
	bot := nyu.GetBot()

	streamer, err := bot.NewEditStreamer(m.Chat, "create editstreamer")
	if err != nil {
		bot.Send(m.Chat, "Ohno, failed to create editstreamer: %s", err)
		log.Printf("Failed to create editstreamer %s", err)

		return
	}

	log := func(f string, a ...any) {
		str := fmt.Sprintf(f, a...) + "\n"

		log.Printf(str)
		streamer.Append(str)
	}

	log("Importing subscriptions from table `subscription`")
	s, err := db.StmtQuery("SELECT count(1) FROM subscription")
	if err != nil {
		log("failed to get count of subscriptions: %s", err)

		return
	}

	var count int64
	if s.Next() {
		err := s.Scan(&count)
		if err != nil {
			log("faield to get count of subs Scan: %s", err)

			return
		}
	}

	s.Close()

	log("Got subscription count: %d", count)

	s, err = db.StmtQuery("SELECT (user_id, type) FROM subscription")
	if err != nil {
		log("failed to get subscriptions: %s", err)

		return
	}

	defer s.Close()

	var user int64
	var sub bool
	for s.Next() {
		err := s.Scan(&user, &sub)
		if err != nil {
			log("faield to scan; skipping this one %s", err)

			return
		}

		if sub { // TODO: figure out which value is correct
			db.SetUserTag(user, status.SpaceStatusSubTag)
		}
	}

	log("Döner!")
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

func handleLeave(m *tele.Message) {
	bot := nyu.GetBot()

	bot.Send(m.Chat, "Goodbye!")

	bot.Leave(m.Chat)
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

func handleTestInline(m *tele.Message) {
	bot := nyu.GetBot()

	// this is very löng
	bot.Send(m.Chat, "asdf",
		&tele.SendOptions{
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: "testinline_ok",

							Text: "ok",
						},
						tele.InlineButton{
							Unique: "testinline_nu",

							Text: "nuu",
						},
					},
				},
			},
		},
	)
}

func handleTestStatusInline(m *tele.Message) {
	status.AskUserIfArrived(m.Sender.ID)
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
