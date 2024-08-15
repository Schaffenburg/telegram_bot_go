package util

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func ParseTime(t string) (newTime time.Time, err error) {
	now := time.Now()

	unix, err := strconv.ParseInt(t, 10, 64)
	if err == nil && unix > now.Unix() {
		return time.Unix(unix, 0), nil
	}

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

func TodayUnknown() time.Time {
	now := time.Now()

	return time.Date(
		now.Year(), now.Month(), now.Day(),
		23, 59, 59, 0, now.Location(),
	)
}

// checks if TIME part of time.Time is 23:59:59 (aka "unknown arival time")
func IsUnknown(t time.Time) bool {
	return t.Hour() == 59 &&
		t.Minute() == 59 &&
		t.Second() == 59
}

var timeFormats []string = []string{
	"15:04",
	"15.04",
	"1504",
	"",
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

func T[K any](c bool, a, b K) K {
	if c {
		return a
	} else {
		return b
	}
}

func ReplaceMulti(m map[string]string, txt string) string {
	for k, v := range m {
		txt = strings.ReplaceAll(txt, k, v)
	}

	return txt
}

func OneOf[K any](r *rand.Rand, of []K) K {
	return of[r.Intn(len(of))]
}

func Join[K any](s []K, delim string) (str string) {
	notFirst := false
	for i := 0; i < len(s); i++ {
		if notFirst {
			str += delim
		}

		str += fmt.Sprintf("%v", s[i])

		notFirst = true
	}

	return
}
