/*
Implementation
*/
package perms

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/stalk"
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
