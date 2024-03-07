package nyu

import (
	tele "gopkg.in/tucnak/telebot.v2"

	colors "github.com/gookit/color"

	"crypto/sha1"
	"log"
	"strconv"
)

func LogMessage(pre string, msg *tele.Message) {
	if msg == nil {
		log.Printf("%s -> *nil*", pre)
	} else {
		log.Printf("%s %s (%s) %s [%s] %s",
			pre,
			ConstColorize(string(msg.Chat.Type)),
			ConstColorize(strconv.FormatInt(msg.Chat.ID, 10)),
			ConstColorize(msg.Sender.Username),
			ConstColorize(strconv.FormatInt(msg.Sender.ID, 10)),
			msg.Text,
		)
	}
}

func LogCallback(pre string, cb *tele.Callback) {
	if cb == nil {
		log.Printf("%s -> *nil*", pre)
	} else {
		log.Printf("%s %s (%s) %s [%s] %s",
			pre,
			ConstColorize(string(cb.Message.Chat.Type)),
			ConstColorize(strconv.FormatInt(cb.Message.Chat.ID, 10)),
			ConstColorize(cb.Sender.Username),
			ConstColorize(strconv.FormatInt(cb.Sender.ID, 10)),
			strconv.Quote(cb.Data),
		)
	}
}

// colorizes str, for any str the colorized output is constant
func ConstColorize(str string) (colorized string) {
	sum := sha1.Sum([]byte(str))

	color := colors.RGB(sum[0], sum[1], sum[2], true)

	return color.Sprint(str)
}
