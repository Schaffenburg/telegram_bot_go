package debug

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/config"
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

var (
	FailGeneric      = loc.MustTrans("fail.generic")
	FailEditstreamer = loc.MustTrans("fail.editstreamer.create")
)

func init() {
	if config.Get().DebugCmd {
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

		bot.Command("debug_gettrans", handleGetTrans, perms...)

		bot.Command("debug_setmembership", handleSetMembership, perms...)
		bot.Command("debug_rmmembership", handleRmMembership, perms...)

		bot.Command("debug_testinlinebtn", handleTestInline, perms...)

		bot.Command("debug_teststatusinline", handleTestStatusInline, perms...)

		bot.Command("debug_leave", handleLeave, perms...)

		bot.Command("debug_importsubscriptions", handleImportSubscriptions, perms...)
		bot.Command("debug_importlangmap", handleImportLanguageMap, perms...)

		bot.Command("debug_dumpusertags", handleDumpUserTags, perms...)
		bot.Command("debug_importstatus", handleImportStatus, perms...)

		debugcallback := func(c *tele.Callback) {
			nyu.GetBot().Reply(c.Message, strconv.Quote(c.Data))
		}

		bot.HandleInlineCallback("testinline_ok", debugcallback)
		bot.HandleInlineCallback("testinline_nu", debugcallback)

		bot.AnswerCommand("hi", "Hi to you too, %u!")
		bot.ReplyCommand("nya", "nya")
	}

	if config.Get().DebugLogInvalidCmd {
		bot := nyu.GetBot()

		// check for any unhandled commands
		bot.Handle(tele.OnText, func(m *tele.Message) {
			if len(m.Text) > 0 && m.Text[0] == '/' {
				bot.Send(m.Chat, "Invalid Command!")
			}
		})
	}

	if config.Get().DebugLog {
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
	}
}

func handleGetTrans(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Sender, "Usage: /debug_gettrans <transid>")

		return
	}

	trans := loc.GetTranslation(args[1])
	if trans == nil {
		bot.Sendf(m.Sender, "unable to get translation for %s", args[1])

		return
	}

	bot.Sendf(m.Sender, "%s in language %s can be translated to %s",
		args[1], l, trans.Get(l))
}

func handleImportLanguageMap(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	streamer, err := bot.NewEditStreamer(m.Chat, "create editstreamer")
	if err != nil {
		bot.Send(m.Chat, FailEditstreamer.Getf(l, err))
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
		1: "Deutsch",
		0: "English",
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

		// check for any unhandled commands
		if config.Get().DebugCmd {
			bot.Handle(tele.OnText, func(m *tele.Message) {
				if len(m.Text) > 0 && m.Text[0] == '/' {
					bot.Send(m.Chat, "Invalid Command!")
				}
			})
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
	l := loc.GetUserLanguage(m.Sender)

	streamer, err := bot.NewEditStreamer(m.Chat, "create editstreamer")
	if err != nil {
		bot.Send(m.Chat, FailEditstreamer.Getf(l, err))
		log.Printf("Failed to create editstreamer %s", err)

		return
	}

	log := func(f string, a ...any) {
		str := fmt.Sprintf(f, a...) + "\n"

		log.Printf(str)
		streamer.Append(str)
	}

	log("Starting spacestatus import from `spacestatus_old`")

	ttl := func() (ttl int) {
		s, err := db.StmtQuery("SELECT count(1) from spacestatus_old;")
		if err != nil {
			log("failed to get spacestatus from DB: %s", err)
		}

		defer s.Close()

		if !s.Next() {
			log("Failed to next ttl: %s", err)

			return -1
		}

		err = s.Scan(&ttl)
		if err != nil {
			log("Failed to scan ttl: %s", err)

			return -1
		}

		return ttl
	}()

	s, err := db.StmtQuery("SELECT status_full, status_timestamp FROM spacestatus_old ORDER BY status_timestamp ASC")
	if err != nil {
		log("failed to get spacestatus from DB: %s", err)
	}

	defer s.Close()

	var status string
	var timestamps string
	var timestamp time.Time
	var errs, ii int = 0, 0
	const batchsize = 20
	var batch = make([][]any, 0, batchsize) //

	for s.Next() {
		if ii%batchsize == 0 || ii > ttl {
			_, err := db.StmtExec("INSERT IGNORE INTO spacestatus (time, status) VALUES (?, ?);",
				batch,
			)

			if err != nil {
				log("Failed to insert: %s", err)
				errs++

				if errs > 10 {
					log("Failed to insert too often")

					return
				}
				continue
			}
			batch = make([][]any, 0, batchsize)

			log("import progress (%d/%d)", ii, ttl)
		}
		ii++

		err = s.Scan(&status, &timestamps)
		if err != nil {
			log("Failed to scan: %s", err)

			return
		}

		timestamp, err = time.Parse(time.DateTime, timestamps)
		if err != nil {
			log("Failed to parse time: %s: %s", timestamps, err)

			continue
		}

		batch = append(batch, []any{timestamp.Unix(), status})
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
	l := loc.GetUserLanguage(m.Sender)

	streamer, err := bot.NewEditStreamer(m.Chat, "create editstreamer")
	if err != nil {
		bot.Send(m.Chat, FailEditstreamer.Getf(l, err))
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

	s, err = db.StmtQuery("SELECT user_id, type FROM subscription")
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

		log("Importing %d type: !%b", user, sub)

		if !sub { // TODO: figure out which value is correct
			db.SetUserTag(user, status.SpaceStatusSubTag)
		}
	}

	log("Döner!")
}

func handleRmTagSelf(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Sender, "Usage: /debug_rmtagself <tag>")

		return
	}

	tag := args[1]

	changed, err := db.RmUserTag(m.Sender.ID, tag)
	if err != nil {
		log.Printf("Error removing user tag: %s", err)
		bot.Send(m.Sender, FailGeneric.Getf(l, err))
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

func handleDumpUserTags(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	var users []int64
	if len(m.Entities) > 0 {
		for _, e := range m.Entities {
			if e.Type == tele.EntityTMention {
				users = append(users, e.User.ID)
			}
		}
	}

	t := &strings.Builder{}

	// only list mentioned users
	if len(users) == 0 {
		var err error
		users, err = db.GetTaggedUsers()
		if err != nil {
			bot.Sendf(m.Chat, FailGeneric.Getf(l, err))
			return
		}
	}

	for _, user := range users {
		u, err := stalk.GetUserByID(user)
		if err != nil {
			bot.Sendf(m.Chat, FailGeneric.Getf(l, err))
			return
		}

		t.WriteString("\n")
		t.WriteString(u.Username)
		t.WriteString(" (")
		t.WriteString(strconv.FormatInt(user, 10))
		t.WriteString("):")

		tags, err := db.GetUserTags(user)
		if err != nil {
			bot.Send(m.Chat, FailGeneric.Getf(l, err))
			return

		}

		for _, tag := range tags {
			t.WriteString("\n - ")
			t.WriteString(tag)
		}
	}

	bot.Send(m.Chat, t.String())
	return
}

func handleDumpGroupTags(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	t, err := db.StmtQuery("SELECT group_id, tag FROM group_tags")
	if err != nil {
		bot.Sendf(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	var group int64
	var tag string

	b := &strings.Builder{}

	for t.Next() {
		err = t.Scan(&group, &tag)
		if err != nil {
			bot.Sendf(m.Chat, FailGeneric.Getf(l, err))

			return
		}

		b.WriteString(strconv.FormatInt(group, 10))
		b.WriteString(" -> ")
		b.WriteString(tag)
		b.WriteString("\n")
	}

	bot.Send(m.Chat, b.String())
}

func handleSetMembership(m *tele.Message) {
	bot := nyu.GetBot()
	//l := loc.GetUserLanguage(m.Sender)

	args := strings.Split(m.Text, " ")
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_setmembership [tgid...] ")

		return
	}

	// remove command
	args = args[1:]

	users := []int64{}
	for _, e := range args {
		// remove ,s from tgs formatting
		clean := strings.ReplaceAll(e, ",", "")

		u, err := strconv.ParseInt(clean, 10, 64)
		if err != nil {
			bot.Sendf(m.Chat, "Failed to decode userid: %s: %s", e, err)
			return
		}

		users = append(users, u)
	}

	var err error
	for _, u := range users {
		err = stalk.UpdateMembership(m.Chat.ID, u)
		if err != nil {
			log.Printf("Failed to update membership for %d group %d: %s", u, m.Chat.ID, err)
		}
	}

	bot.Sendf(m.Chat, "Added Group Memberships for group %d for users %v", m.Chat.ID, users)
}

func handleRmMembership(m *tele.Message) {
	bot := nyu.GetBot()
	//l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_rmmembership [tgid...] ")

		return
	}

	// remove command
	args = args[1:]

	users := []int64{}
	for _, e := range args {
		// remove ,s from tgs formatting
		clean := strings.ReplaceAll(e, ",", "")

		u, err := strconv.ParseInt(clean, 10, 64)
		if err != nil {
			bot.Sendf(m.Chat, "Failed to decode userid: %s: %s", e, err)
			return
		}

		users = append(users, u)
	}

	var err error
	for _, u := range users {
		_, err = stalk.UpdateMembershipDelete(m.Chat.ID, u)
		if err != nil {
			log.Printf("Failed to update membership for %d group %d: %s", u, m.Chat.ID, err)
		}
	}

	bot.Sendf(m.Chat, "Removed Group Memberships for group %d for users %v", m.Chat.ID, users)
}

func handleSetTagSelf(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_settagself <tag>")

		return
	}

	tag := args[1]

	err := db.SetUserTag(m.Sender.ID, tag)
	if err != nil {
		log.Printf("Error adding user tag: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, "Set tag.")
	}
}

func handleIsGroupMemberSelf(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_isgroupmemberself <group(int64)>")

		return
	}

	group, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	}

	status, err := stalk.IsMember(m.Sender.ID, group)
	if err != nil {
		log.Printf("Error checking membership: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, "You are"+util.T(status, "", "n't")+" a member.")
	}
}

func handleSetGroupTagCurrent(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	if m.Chat.Type == tele.ChatPrivate {
		bot.Send(m.Chat, "grouptags are not available for Chat.Type \"ChatPrivate\".")
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
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
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
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) < 2 {
		bot.Send(m.Chat, "Usage: /debug_istaggedgroupmemberself <tag[len:20]>")

		return
	}

	tag := args[1]

	status, err := stalk.IsTaggedGroupMember(m.Sender.ID, tag)
	if err != nil {
		log.Printf("Error getting membership status of tagged group: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, "You are"+util.T(status, "", "n't")+" a member.")
	}
}
