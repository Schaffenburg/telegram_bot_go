package pager

import (
	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
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

		for {
			status = <-updateCh

			text := "[nyla] " + status.Text()

			log.Printf("pager send '%s'", text)
			ids, err := ListBroadcastIDs()
			if err != nil {
				log.Printf("Error listing pager broadcasting IDs %s", err)
				continue
			}

			for i := 0; i < len(ids); i++ {
				log.Printf("Sending emsg to %d", ids[i])

				res, err := emsg.SendMessage(strconv.FormatInt(ids[i], 10), text)
				if err != nil || res.Status != "success" {
					log.Printf("Failed to send emsg status to %d: %s", ids[i], err)

					// TODO: notify ownder of pager
					ownerID, err := GetPagerIDOwner(ids[i])
					if err != nil {
						continue
					}
					owner := &tele.User{ID: ownerID}

					if res == nil {
						bot.Sendf(owner, "Dein Pager %d konnte nicht benachichtigt werden: error %s", ids[i], err)
					} else {
						bot.Sendf(owner, "Dein Pager %d konnte nicht benachichtigt werden: status %s", ids[i], res.Status)
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
		bot.Sendf(m.Chat, "Ohno, %s", err)

		return
	}

	if !ch {
		bot.Send(m.Chat, "Diese eID ist schon in verwendung")

		return
	}

	bot.Send(m.Chat, "Ok, ist hinzugefuegt")
}

func handleRmPager(m *tele.Message) {
	bot := nyu.GetBot()

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
		bot.Sendf(m.Chat, "Ohno, %s", err)

		return
	}

	if !ch {
		bot.Send(m.Chat, "Wusste gar nicht, dass du diese eID hattest")

		return
	}

	bot.Send(m.Chat, "Ok, ist entfernt")
}

func handleListPager(m *tele.Message) {
	bot := nyu.GetBot()

	pagers, err := ListPagers(m.Sender.ID)
	if err != nil {
		log.Printf("Failed to list pager for user %d: %s", m.Sender.ID, err)
		bot.Sendf(m.Chat, "Ohno, %s", err)

		return
	}

	if len(pagers) == 0 {
		bot.Send(m.Chat, "Du hast noch keine Pager hinzugefuegt")

		return
	}

	bot.Send(m.Chat, "Du hast folgende IDs: "+util.Join(pagers, ", "))
}
