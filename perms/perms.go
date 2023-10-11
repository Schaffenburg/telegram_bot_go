/*
Implementation
*/
package perms

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/stalk"

	"strings"
)

var (
	GroupEV  = &PermissionGroupTag{"perm_ev"}
	GroupCIX = &PermissionGroupTag{"perm_cix"}

	// if member of any space releated group
	MemberSpaceGroup = &PermissionOr{
		GroupCIX,
		GroupEV,
	}
)

// specifies permission of which one must be satisfied
type PermissionOr []Permission

func (p PermissionOr) Check(m *tele.Message) (ok bool, err error) {
	for _, p := range p {
		ok, err = p.Check(m)
		if ok {
			return true, nil
		}
	}

	return ok, nil
}

func (p PermissionOr) String() string {
	b := &strings.Builder{}

	b.WriteString("(")

	notFirst := false
	for _, p := range p {
		if notFirst {
			b.WriteString(" or ")
		}

		b.WriteString(p.String())
	}

	b.WriteString(")")

	return b.String()
}

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
