package perms

import (
	tele "gopkg.in/tucnak/telebot.v2"
	"log"

	db "git.schaffenburg.org/nyu/schaffenbot/database"
	"git.schaffenburg.org/nyu/schaffenbot/nyu"
	"git.schaffenburg.org/nyu/schaffenbot/stalk"
)

type PermissionTag struct {
	Tag string
}

func (p *PermissionTag) String() string {
	return "Tag(" + p.Tag + ")"
}

func (p *PermissionTag) Check(m *tele.Message) (bool, error) {
	return db.UserHasTag(m.Sender.ID, p.Tag)
}

type PermissionGroupTag struct {
	GroupTag string
}

func (p *PermissionGroupTag) Check(m *tele.Message) (bool, error) {
	return stalk.IsTaggedGroupMember(m.Sender.ID, p.GroupTag)
}

func (p *PermissionGroupTag) String() string {
	return "Member of group with GroupTag(" + p.GroupTag + ")"
}

type Permission interface {
	Check(*tele.Message) (bool, error)

	String() string
}

func Require(f func(*tele.Message), perms ...Permission) func(*tele.Message) {
	return func(m *tele.Message) {
		bot := nyu.Bot()

		var ok bool
		var err error

		for _, perm := range perms {
			ok, err = perm.Check(m)
			if err != nil {
				log.Printf("Permission check failed: %s", err)
				bot.Send(m.Chat, "Rechte ueberpruefung fehlgeschlagen!")

				return
			}

			if !ok {
				bot.Send(m.Chat, "Folgende Anforderung wird nicht erfuellt: "+perm.String())

				return
			}
		}

		f(m)
	}
}
