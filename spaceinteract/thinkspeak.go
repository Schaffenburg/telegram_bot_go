package interact

import (
	"encoding/json"
	"github.com/Schaffenburg/telegram_bot_go/config"
	"net/http"
	"strconv"

	"log"
	"time"
)

type ThinkspeakResponse struct {
	Channel Channel `json:"channel"`
	Feeds   []Feed  `json:"feeds"`
}

type Channel struct {
	hahanometolazy string
}

type Time time.Time

func (t *Time) UnmarshalJSON(b []byte) error {
	var buf string
	err := json.Unmarshal(b, &buf)
	if err != nil {
		return err
	}

	ti, err := time.Parse("2006-01-02T15:04:05Z", buf)
	if err != nil {
		return err
	}

	*t = Time(ti)

	return nil
}

type Feed struct {
	CreatedAt Time        `json:"created_at"`
	EntryId   int         `json:"entry_id"`
	Field2    FloatString `json:"field2"`
}

type FloatString float32

func (fs *FloatString) UnmarshalJSON(b []byte) error {
	var buf string
	err := json.Unmarshal(b, &buf)
	if err != nil {
		return err
	}

	f64, err := strconv.ParseFloat(buf, 10)
	if err != nil {
		return err
	}

	*fs = FloatString(float32(f64))

	return nil
}

func GetTemp() (temp float32, measured time.Time, err error) {
	resp, err := http.Get(config.Get().ThinkSpeakTemp)
	if err != nil {
		return 0, time.Date(0, 0, 0, 0, 0, 0, 0, nil), err
	}

	dec := json.NewDecoder(resp.Body)
	re := new(ThinkspeakResponse)
	err = dec.Decode(&re)
	if err != nil {
		return 0, time.Date(0, 0, 0, 0, 0, 0, 0, nil), err
	}

	log.Printf("%+#v", re)

	log.Printf("Temperature is: %.1f °C (measured at %s)",
		re.Feeds[0].Field2, time.Time(re.Feeds[0].CreatedAt).String())

	return float32(re.Feeds[0].Field2), time.Time(re.Feeds[0].CreatedAt), nil
}
