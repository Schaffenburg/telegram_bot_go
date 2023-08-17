package nyu

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"git.schaffenburg.org/nyu/schaffenbot/config"
	"git.schaffenburg.org/nyu/schaffenbot/database"

	"log"
	"sync"
	"time"
)

var (
	poller  *ProxyPoller
	bot     *tele.Bot
	botOnce sync.Once
)

func Poller() *ProxyPoller {
	botOnce.Do(makeBot)

	return poller
}

func Bot() *tele.Bot {
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

	var err error
	bot, err = tele.NewBot(s)
	if err != nil {
		log.Fatalf("Error adding bot: %s\n", err)
	}
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
	log.SetFlags(log.Flags() | log.Lshortfile)

	bot := Bot()

	database.DB()

	log.Printf("starting telebot")
	handleRun()
	bot.Start()
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
