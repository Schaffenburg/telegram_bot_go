package loc

import (
	db "github.com/Schaffenburg/telegram_bot_go/database"
	tele "gopkg.in/tucnak/telebot.v2"

	"bufio"
	"embed"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"strings"
	"sync"
)

const DefaultLanguage = "Deutsch"

//go:embed locale
var localizeFS embed.FS

const localizeFSprefix = "locale"

var languagesLookup = make(map[string]*Language)
var uniqueLanguages = make([]*Language, 0)
var translationsLookup = make(map[string]int)

// map language textid -> translation
var translations = make(map[int]map[int]string)
var translationsInc = 0
var translationsOnce sync.Once

func init() {
	_, err := db.StmtExec("CREATE TABLE IF NOT EXISTS `language` (`user` BIGINT NOT NULL PRIMARY KEY, `language` TEXT);")
	if err != nil {
		log.Fatalf("Failed to create table 'language': %s", err)
	}
}

func SetUserLanguage(u int64, l Language) error {
	_, err := db.StmtExec("INSERT INTO language (user, language) VALUES (?, ?) ON DUPLICATE KEY UPDATE language = VALUES(language)",
		u, l.name,
	)

	return err
}

func initLocale() {
	translationsOnce.Do(func() {
		d, err := localizeFS.ReadDir(localizeFSprefix)
		if err != nil {
			log.Fatalf("Failed to read localizeFS '%s': %s", localizeFSprefix, err)
		}

		for i := 0; i < len(d); i++ {
			if d[i].IsDir() {
				continue
			}

			path := localizeFSprefix + "/" + d[i].Name()
			f, err := localizeFS.Open(path)
			if err != nil {
				log.Fatalf("Failed to open localizeFS/ %s: %s", path, err)
			}

			r := bufio.NewReader(f)
			b, err := r.ReadBytes('\n')
			languages := strings.Split(string(b), ",")

			langmap := make(map[string]string)
			d := yaml.NewDecoder(r)

			err = d.Decode(&langmap)
			if err != nil {
				log.Fatalf("Failed to decode languagemap %s: %s", path, err)
			}

			lang := &Language{id: i, name: languages[0], iso: languages[1]}
			for lI := 0; lI < len(languages); lI++ {
				languagesLookup[languages[lI]] = lang // set language to outer i
			}

			uniqueLanguages = append(uniqueLanguages, lang)

			for name, trans := range langmap {
				// get translations id
				id, ok := translationsLookup[name]
				if !ok {
					id = translationsInc
					translationsInc++

					translationsLookup[name] = id
				}

				_, ok = translations[id]
				if !ok {
					translations[id] = make(map[int]string)
				}

				translations[id][i] = trans // id = translation ID; i is language ID
			}
		}
	})
}

func MustTrans(name string) Translation {
	t := GetTranslation(name)
	if t == nil {
		log.Fatalf("Cant get translation for %s!", name)
	}

	return *t
}

func GetTranslation(name string) *Translation {
	initLocale()

	id, ok := translationsLookup[name]
	if !ok {
		return &Translation{
			name: name,
			id:   -1,
		}
	}

	return &Translation{
		name: name,
		id:   id,
	}
}

func GetLanguage(name string) *Language {
	initLocale()

	l, ok := languagesLookup[name]
	if !ok {
		return nil
	}

	return l
}

func DefaultLang() Language {
	l := GetLanguage(DefaultLanguage)
	if l == nil {
		log.Fatalf("Default language %s not defined!", DefaultLanguage)
	}

	return *l
}

type Translation struct {
	name string
	id   int
}

func (t *Translation) Name() string {
	return t.name
}

type Language struct {
	name string
	iso  string
	id   int
}

func (t *Language) Name() string {
	return t.name
}

func (t *Language) ISO() string {
	return t.iso
}

func (t *Language) ID() int {
	return t.id
}

// Gets the translation for a language
func (t Translation) Get(l *Language) string {
	if l == nil {
		return t.name
	}

	lang, ok := translations[t.id]
	if !ok {
		return t.name
	}

	trans, ok := lang[l.id]
	if !ok {
		return t.name
	}

	return trans
}

// Gets the translation for a language
func (t Translation) Getf(l *Language, a ...any) string {
	trans := t.Get(l)

	return fmt.Sprintf(trans, a...)
}

func SetUserLanguageAuto(u int64) error {
	_, err := db.StmtExec("DELETE FROM language WHERE user = ?", u)

	return err
}

func GetUserLanguage(u *tele.User) *Language {
	l, _ := GetUserLanguageV(u)

	return l
}

// language can be nil (by user ID)
func GetUserLanguageID(id int64) (l *Language) {
	r, err := db.StmtQuery("SELECT language FROM language WHERE user = ?", id)
	if err != nil {
		return nil
	}

	if !r.Next() {
		return nil
	}

	var lang string
	err = r.Scan(&lang)
	if err != nil {
		log.Printf("Failed to scan langauge from user %d: %s", id, err)

		return nil
	}

	return GetLanguage(lang)
}

// returns default language if user has nothing configured
func MustGetUserLanguageID(id int64) (l *Language) {
	l = GetUserLanguageID(id)
	if l == nil {
		return
	}

	return GetLanguage(DefaultLanguage)
}

// verbose also returns if is auto
func GetUserLanguageV(u *tele.User) (l *Language, auto bool) {
	auto = false

	code := GetUserLanguageID(u.ID)

	if code == nil && u.LanguageCode != "" {
		auto = true
		code = GetLanguage(u.LanguageCode)
	}

	if code == nil {
		auto = false
		l = GetLanguage(DefaultLanguage)
	}

	return
}

func GetLanguagesS() (s []string) {
	s = make([]string, len(uniqueLanguages))

	for i, lang := range uniqueLanguages {
		s[i] = lang.name
	}

	return
}

// Returns list of all unique languages; pls treat array as read only
func GetLanguages() []*Language {
	return uniqueLanguages
}
