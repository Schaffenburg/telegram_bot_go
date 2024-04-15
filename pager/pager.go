package pager

import (
	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/status"
	"github.com/Schaffenburg/telegram_bot_go/util"
	tele "gopkg.in/tucnak/telebot.v2"

	"errors"
	emsg "github.com/derzombiiie/emessage"
	"log"
	"strconv"
	"strings"
)

var (
	ErrNoEntries = errors.New("No entries found")
)

var (
	LNotifyHTTP    = loc.MustTrans("pager.notify.http")
	LNotifyAPI     = loc.MustTrans("pager.notify.api")
	LeIDTaken      = loc.MustTrans("pager.eID.add.taken")
	LeIDAddConfirm = loc.MustTrans("pager.eID.add.confirm")
	LeIDrmnochange = loc.MustTrans("pager.eID.rm.nochange")
	LeIDremoved    = loc.MustTrans("pager.eID.rm.confirm")
	Lnopagers      = loc.MustTrans("pager.list.nopagers")
	Lpagerlist     = loc.MustTrans("pager.list")

	GenericFail = loc.MustTrans("fail.generic")
)

func init() {
	updateCh := make(chan status.SpaceStatus)
	status.AddStatusCh(updateCh)

	err := db.StartDB()
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err)
	}

	database := db.DB()

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `pager_broadcast` ( `user`BIGINT , `eID` BIGINT PRIMARY KEY );")
	if err != nil {
		log.Printf("error creating table arrivalTimes: %s", err)
		return
	}

	bot := nyu.GetBot()

	bot.Command("pagerhinzufuegen", handleAddPager)
	help.AddCommand("pagerhinzufuegen")

	bot.Command("pagerentfernen", handleRmPager)
	help.AddCommand("pagerentfernen")

	bot.Command("listpagers", handleListPager)
	help.AddCommand("listpagers")

	go func() {
		var status status.SpaceStatus
		var text string

		for {
			status = <-updateCh

			log.Printf("pager send '%s'", text)
			ids, err := ListBroadcastIDs()
			if err != nil {
				log.Printf("Error listing pager broadcasting IDs %s", err)
				continue
			}

			for i := 0; i < len(ids); i++ {
				log.Printf("Sending emsg to %d", ids[i])

				// TODO: notify ownder of pager
				ownerID, err := GetPagerIDOwner(ids[i])
				if err != nil {
					continue
				}
				l := loc.MustGetUserLanguageID(ownerID)

				text = "[nyla] " + status.Text(l)

				res, err := emsg.SendMessage(strconv.FormatInt(ids[i], 10), text)
				if err != nil || res.Status != "success" {
					log.Printf("Failed to send emsg status to %d: %s", ids[i], err)

					owner := nyu.Recipient(ownerID)

					if res == nil {
						bot.Send(owner, LNotifyHTTP.Getf(l, ids[i], err))
					} else {
						bot.Send(owner, LNotifyAPI.Getf(l, ids[i], res.Status))
					}
				}
			}
		}
	}()
}

func GetPagerIDOwner(eID int64) (user int64, err error) {
	res, err := db.StmtQuery("SELECT user FROM pager_broadcast WHERE eID = ?", eID)
	if err != nil {
		return 0, err
	}

	if !res.Next() {
		return 0, ErrNoEntries
	}

	err = res.Scan(&user)
	if err != nil {
		return 0, err
	}

	return
}

func ListBroadcastIDs() (s []int64, err error) {
	res, err := db.StmtQuery("SELECT eID FROM pager_broadcast")
	if err != nil {
		return nil, err
	}

	var eID int64
	for res.Next() {
		err = res.Scan(&eID)
		if err != nil {
			return nil, err
		}

		s = append(s, eID)
	}

	return
}

func AddPagerID(user, pager int64) (changed bool, err error) {
	r, err := db.StmtExec("INSERT INTO pager_broadcast (user, eID) VALUES (?, ?);",
		user, pager,
	)
	if err != nil {
		return false, err
	}

	i, err := r.RowsAffected()
	return i != 0, err
}

func RmPagerID(user, pager int64) (changed bool, err error) {
	r, err := db.StmtExec("DELETE FROM pager_broadcast WHERE user = ? AND eID = ?;",
		user, pager,
	)
	if err != nil {
		return false, err
	}

	i, err := r.RowsAffected()
	return i != 0, err
}

func ListPagers(user int64) (s []int64, err error) {
	r, err := db.StmtQuery("SELECT eID FROM pager_broadcast WHERE user = ?;",
		user,
	)
	if err != nil {
		return nil, err
	}

	var eID int64
	for r.Next() {
		err = r.Scan(&eID)
		if err != nil {
			return nil, err
		}

		s = append(s, eID)
	}

	return
}

func handleAddPager(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage /pagerhinzufuegen <eid [int64]>")
		return
	}

	eid, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		bot.Send(m.Chat, "Usage /pagerhinzufuegen <eid [int64]>")
		return
	}

	ch, err := AddPagerID(m.Sender.ID, eid)
	if err != nil {
		log.Printf("Failed to add pager for user %d: %s", m.Sender.ID, err)
		bot.Sendf(m.Chat, GenericFail.Getf(l, err))

		return
	}

	if !ch {
		bot.Send(m.Chat, LeIDTaken.Get(l))

		return
	}

	bot.Send(m.Chat, LeIDAddConfirm.Get(l))
}

func handleRmPager(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage /pagerentfernen <eid [int64]>")
		return
	}

	eid, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		bot.Send(m.Chat, "Usage /pagerentfernen <eid [int64]>")
		return
	}

	ch, err := RmPagerID(m.Sender.ID, eid)
	if err != nil {
		log.Printf("Failed to rm pager for user %d: %s", m.Sender.ID, err)
		bot.Sendf(m.Chat, GenericFail.Getf(l, err))

		return
	}

	if !ch {
		bot.Send(m.Chat, LeIDrmnochange.Get(l))
	} else {
		bot.Send(m.Chat, LeIDremoved.Get(l))
	}
}

func handleListPager(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	pagers, err := ListPagers(m.Sender.ID)
	if err != nil {
		log.Printf("Failed to list pager for user %d: %s", m.Sender.ID, err)
		bot.Sendf(m.Chat, GenericFail.Getf(l, err))

		return
	}

	if len(pagers) == 0 {
		bot.Send(m.Chat, Lnopagers.Get(l))

		return
	}

	bot.Send(m.Chat, Lpagerlist.Getf(l, util.Join(pagers, ", ")))
}
