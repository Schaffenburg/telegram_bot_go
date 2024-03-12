package nyu

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/config"
	"github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/util"

	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Recipient int64

func (r Recipient) Recipient() string {
	return strconv.FormatInt(int64(r), 10)
}

type Bot struct {
	*tele.Bot

	callbackLookupMu sync.RWMutex
	callbackLookup   map[string]func(q *tele.Callback)
}

func (bot *Bot) RespondText(c *tele.Callback, text string) error {
	return bot.Respond(c, &tele.CallbackResponse{Text: text})
}

func (bot *Bot) GetCurrentPFP(u *tele.User) (*tele.Photo, error) {
	p, err := bot.ProfilePhotosOf(u)
	if err != nil {
		return nil, err
	}

	return &p[0], nil
}

func (b *Bot) Command(command string, h func(m *tele.Message), perms ...Permission) {
	b.Bot.Handle("/"+command, handlePermit(h, perms...))
}

func (b *Bot) HandleInlineCallback(unique string, f func(q *tele.Callback)) {
	b.callbackLookupMu.Lock()
	defer b.callbackLookupMu.Unlock()

	if b.callbackLookup == nil {
		b.callbackLookup = make(map[string]func(*tele.Callback))
	}

	b.callbackLookup["\f"+unique] = f
}

// should not be used for Commands! please also try to use OnText/OnVideo/etc. functions instead!
func (b *Bot) Handle(command string, h func(m *tele.Message), perms ...Permission) {
	b.Bot.Handle(command, handlePermit(h, perms...))
}

//func (b *Bot) Send(destination tele.Recipient, text string) (*tele.Message, error) {
//	return b.Bot.Send(destination, text)
//}

// options can be nil
func (b *Bot) Sendf(destination tele.Recipient, format string, args ...any) (*tele.Message, error) {
	return b.Bot.Send(destination, fmt.Sprintf(format, args...))
}

func (b *Bot) Replyf(to *tele.Message, format string, args ...any) (*tele.Message, error) {
	return b.Bot.Reply(to, fmt.Sprintf(format, args...))
}

func (b *Bot) AnswerCommand(command, text string, perms ...Permission) {
	b.Command(command, func(m *tele.Message) {
		bot.Send(m.Chat, util.ReplaceMulti(map[string]string{
			"%u": m.Sender.FirstName,
			"%t": m.Text,
			"%h": m.Sender.Username,
		}, text))
	}, perms...)
}

func (b *Bot) ReplyCommand(command, text string, perms ...Permission) {
	b.Command(command, func(m *tele.Message) {
		bot.Reply(m, util.ReplaceMulti(map[string]string{
			"%u": m.Sender.FirstName,
			"%t": m.Text,
			"%h": m.Sender.Username,
		}, text))
	}, perms...)
}

var (
	poller  *ProxyPoller
	bot     *Bot
	botOnce sync.Once
)

func Poller() *ProxyPoller {
	botOnce.Do(makeBot)

	return poller
}

func GetBot() *Bot {
	botOnce.Do(makeBot)

	return bot
}

func makeBot() {
	conf := config.Get()

	poller = &ProxyPoller{
		Poller: &tele.LongPoller{
			Limit:   100,
			Timeout: time.Second,
		},
	}

	s := tele.Settings{
		Token: conf.Token,

		Poller: poller,
	}

	b, err := tele.NewBot(s)
	if err != nil {
		log.Fatalf("Error adding bot: %s\n", err)
	}

	bot = &Bot{Bot: b}

	bot.Bot.Handle(tele.OnCallback, func(c *tele.Callback) {
		bot.callbackLookupMu.RLock()
		defer bot.callbackLookupMu.RUnlock()

		f, ok := bot.callbackLookup[c.Data]
		if ok {
			f(c)
		}
	})
}

type ProxyPoller struct {
	Poller tele.Poller
	sync.RWMutex

	updateChs []chan tele.Update
}

func (p *ProxyPoller) AddCh(ch chan tele.Update) {
	p.Lock()
	defer p.Unlock()

	p.updateChs = append(p.updateChs, ch)
}

func (p *ProxyPoller) Poll(b *tele.Bot, updates chan tele.Update, stop chan struct{}) {
	p.AddCh(updates)

	uch := make(chan tele.Update, 100)
	go func() {
		var u tele.Update

		for {
			select {
			case u = <-uch:
				p.RLock()
				// start with last addition
				for i := len(p.updateChs) - 1; i >= 0; i-- {
					p.updateChs[i] <- u
				}

				p.RUnlock()

			case <-stop:
				return
			}
		}
	}()

	p.Poller.Poll(b, uch, stop)
}

func Run() {
	log.SetFlags(log.Flags() | log.Lshortfile) // log.Llongfile) // |

	bot := GetBot()

	database.DB()

	log.Printf("starting telebot")
	handleRun()

	bot.Command("start", handleStart)

	bot.Bot.Start()
}

func handleStart(m *tele.Message) {
	bot := GetBot()

	bot.Send(m.Sender, "Hi, ich bin nyu, der Schaffenburg bot\nmir kannst du gerne ankuendigen, wenn du vor hast in den space zu kommen.\nich kann auch viel misc zeugs, bei interesse kannst du dir gerne mein /help durchlesen ^^")
}

var (
	runFuncs   []func()
	runFuncsMu sync.Mutex
)

func OnRun(f func()) {
	runFuncsMu.Lock()
	defer runFuncsMu.Unlock()

	runFuncs = append(runFuncs, f)
}

func handleRun() {
	runFuncsMu.Lock()
	defer runFuncsMu.Unlock()

	for i := 0; i < len(runFuncs); i++ {
		runFuncs[i]()
	}
}

func (bot *Bot) NewEditStreamer(chat *tele.Chat, text string) (*EditStreamer, error) {
	return NewEditStreamer(bot, chat, text)
}

func NewEditStreamer(bot *Bot, chat *tele.Chat, text string) (*EditStreamer, error) {
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

	bot *Bot

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
