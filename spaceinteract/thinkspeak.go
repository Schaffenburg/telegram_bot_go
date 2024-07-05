package interact

import (
	"encoding/json"
	"github.com/Schaffenburg/telegram_bot_go/config"
	"net/http"
	"strconv"

	"log"
)

type ThinkspeakResponse struct {
	Channel Channel `json:"channel"`
	Feeds   []Feed  `json:"feeds"`
}

type Channel struct {
	hahanometolazy string
}

type Feed struct {
	CreatedAt string      `json:"created_at"`
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

func GetTemp() (float32, error) {
	resp, err := http.Get(config.Get().ThinkSpeakTemp)
	if err != nil {
		return 0, err
	}

	dec := json.NewDecoder(resp.Body)
	re := new(ThinkspeakResponse)
	err = dec.Decode(&re)
	if err != nil {
		return 0, err
	}

	log.Printf("%+#v", re)

	log.Printf("Temperature is: %.1f °C", re.Feeds[0].Field2)

	return float32(re.Feeds[0].Field2), nil
}
