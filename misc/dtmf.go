package misc

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/nyu"

	"sync"
	"time"
)

const (
	CallbackDTMF1 = "misc_dtmf_1"
	CallbackDTMF2 = "misc_dtmf_2"
	CallbackDTMF3 = "misc_dtmf_3"
	CallbackDTMF4 = "misc_dtmf_4"
	CallbackDTMF5 = "misc_dtmf_5"
	CallbackDTMF6 = "misc_dtmf_6"
	CallbackDTMF7 = "misc_dtmf_7"
	CallbackDTMF8 = "misc_dtmf_8"
	CallbackDTMF9 = "misc_dtmf_9"
	CallbackDTMF0 = "misc_dtmf_0"

	CallbackDTMFH = "misc_dtmf_H" // hash
	CallbackDTMFS = "misc_dtmf_S" // star
)

type DTMFthing struct {
	*nyu.EditStreamer

	Timeout time.Time
}

var (
	dtmfMap   map[int64]*DTMFthing = make(map[int64]*DTMFthing)
	dtmfMapMu sync.Mutex
)

func init() {
	bot := nyu.GetBot()

	bot.Command("dtmf", ShowDTMF)
	help.AddCommand(tele.Command{
		Text:        "dtmf",
		Description: "show DTMF keyboard",
	})

	bot.HandleInlineCallback(CallbackDTMF1, callbackKeyboard("1"))
	bot.HandleInlineCallback(CallbackDTMF2, callbackKeyboard("2"))
	bot.HandleInlineCallback(CallbackDTMF3, callbackKeyboard("3"))
	bot.HandleInlineCallback(CallbackDTMF4, callbackKeyboard("4"))
	bot.HandleInlineCallback(CallbackDTMF5, callbackKeyboard("5"))
	bot.HandleInlineCallback(CallbackDTMF6, callbackKeyboard("6"))
	bot.HandleInlineCallback(CallbackDTMF7, callbackKeyboard("7"))
	bot.HandleInlineCallback(CallbackDTMF8, callbackKeyboard("8"))
	bot.HandleInlineCallback(CallbackDTMF9, callbackKeyboard("9"))
	bot.HandleInlineCallback(CallbackDTMF0, callbackKeyboard("0"))

	bot.HandleInlineCallback(CallbackDTMFH, callbackKeyboard("#"))
	bot.HandleInlineCallback(CallbackDTMFS, callbackKeyboard("*"))
}

func callbackKeyboard(char string) func(*tele.Callback) {

	return func(c *tele.Callback) {
		bot := nyu.GetBot()

		dtmfMapMu.Lock()
		defer dtmfMapMu.Unlock()

		t, ok := dtmfMap[c.Message.Chat.ID]
		if !ok || time.Now().After(t.Timeout) {
			es, err := bot.NewEditStreamer(c.Message.Chat, "Dialing: ")
			if err != nil {
				bot.Send(c.Message.Chat, "Ohno, "+err.Error())

				return
			}

			t = &DTMFthing{
				EditStreamer: es,

				Timeout: time.Now().Add(time.Second * 10),
			}

			dtmfMap[c.Message.Chat.ID] = t
		}

		t.Append(char)
		t.Timeout = time.Now().Add(time.Second * 10)

		bot.RespondText(c, char)
	}
}

func ShowDTMF(m *tele.Message) {
	SendDTMF(m.Chat)
}

func SendDTMF(r tele.Recipient) {
	nyu.GetBot().Send(r, "Telefon",
		&tele.SendOptions{
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDTMF1,

							Text: "1",
						},
						tele.InlineButton{
							Unique: CallbackDTMF2,

							Text: "2",
						},
						tele.InlineButton{
							Unique: CallbackDTMF3,

							Text: "3",
						},
					},
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDTMF4,

							Text: "4",
						},
						tele.InlineButton{
							Unique: CallbackDTMF5,

							Text: "5",
						},
						tele.InlineButton{
							Unique: CallbackDTMF6,

							Text: "6",
						},
					},
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDTMF7,

							Text: "7",
						},
						tele.InlineButton{
							Unique: CallbackDTMF8,

							Text: "8",
						},
						tele.InlineButton{
							Unique: CallbackDTMF9,

							Text: "9",
						},
					},
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDTMFH,

							Text: "#",
						},
						tele.InlineButton{
							Unique: CallbackDTMF0,

							Text: "0",
						},
						tele.InlineButton{
							Unique: CallbackDTMFS,

							Text: "*",
						},
					},
				},
			},
		},
	)
}
