package util

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"encoding/json"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ParseTime(t string) (newTime time.Time, err error) {
	now := time.Now()

	for _, fmt := range timeFormats {
		newTime, err = time.Parse(fmt, t)
		if err != nil {
			continue
		} else {
			break
		}
	}

	if err != nil {
		return
	}

	newTime = time.Date(
		now.Year(), now.Month(), now.Day(),
		newTime.Hour(), newTime.Minute(), 0, 0, now.Location(),
	)

	return newTime, nil
}

func must[A any, R any](f func(A) (R, error), a A) (r R) {
	r, err := f(a)
	if err != nil {
		panic(err)
	}

	return r
}

func Today(d time.Duration) time.Time {
	now := time.Now()

	return time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location(),
	).Add(d)
}

var timeFormats []string = []string{
	"15:04",
	"15.04",
	"1504",
}

func TimeFormats() []string {
	s := make([]string, len(timeFormats))

	copy(s, timeFormats)
	return s
}

func GetFullUser(b *tele.Bot, id int64) (*tele.User, error) {
	params := map[string]string{
		"id": strconv.FormatInt(id, 10),
	}

	data, err := b.Raw("usersjgetFullUser", params)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result *tele.User
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return resp.Result, nil

}

func NewEditStreamer(bot *tele.Bot, chat *tele.Chat, text string) (*EditStreamer, error) {
	es := new(EditStreamer)
	es.buf = &strings.Builder{}
	es.bot = bot

	var err error
	es.Message, err = es.bot.Send(chat, text)
	if err != nil {
		return nil, err
	}

	es.buf.WriteString(text)

	return es, nil
}

type EditStreamer struct {
	*tele.Message
	sync.Mutex

	bot *tele.Bot

	buf *strings.Builder
}

func (es *EditStreamer) Append(s string) error {
	es.Lock()
	defer es.Unlock()

	es.buf.WriteString(s)

	// send update
	m, err := es.bot.Edit(es.Message, es.buf.String())
	if err != nil {
		return err
	}

	es.Message = m

	return nil
}

func T[K any](c bool, a, b K) K {
	if c {
		return a
	} else {
		return b
	}
}

func AnswerHandler(bot *tele.Bot, txt string) func(m *tele.Message) {
	return func(m *tele.Message) {
		bot.Send(m.Chat, ReplaceMulti(map[string]string{
			"%u": m.Sender.FirstName,
			"%t": m.Text,
			"%h": m.Sender.Username,
		}, txt))
	}
}

func ReplyHandler(bot *tele.Bot, txt string) func(m *tele.Message) {
	return func(m *tele.Message) {
		bot.Reply(m, ReplaceMulti(map[string]string{
			"%u": m.Sender.FirstName,
			"%t": m.Text,
			"%h": m.Sender.Username,
		}, txt))
	}
}

func ReplaceMulti(m map[string]string, txt string) string {
	for k, v := range m {
		txt = strings.ReplaceAll(txt, k, v)
	}

	return txt
}

func GetCurrentPFP(b *tele.Bot, u *tele.User) (*tele.Photo, error) {
	p, err := b.ProfilePhotosOf(u)
	if err != nil {
		return nil, err
	}

	return &p[0], nil
}

func OneOf[K any](r *rand.Rand, of []K) K {
	return of[r.Intn(len(of))]
}
