package database

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/Schaffenburg/telegram_bot_go/config"
	"github.com/Schaffenburg/telegram_bot_go/util"

	_ "github.com/go-sql-driver/mysql"
)

var (
	database     *sql.DB
	databaseOnce sync.Once
)

func DB() *sql.DB {
	ensureDB()

	return database
}

func openDB() (err error) {
	conf := config.Get()

	database, err = sql.Open(conf.DBDriver, conf.DBSource)
	if err != nil {
		return
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `memberships` (`group_id` bigint(20) NOT NULL, `user` bigint(20) NOT NULL, PRIMARY KEY (`group_id`,`user`))")
	if err != nil {
		log.Printf("error creating table memberships: %s", err)
		return
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `group_tags` ( `group_id` BIGINT NOT NULL, `tag` VARCHAR(20) NOT NULL, PRIMARY KEY (`group_id`, `tag`))")
	if err != nil {
		log.Printf("error creating table group_tags: %s", err)
		return
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `log` ( `user` BIGINT NOT NULL, `time` BIGINT NOT NULL, `key` VARCHAR(20) NOT NULL, `value` TEXT NOT NULL, PRIMARY KEY (`user`, `time`, `key`))")
	if err != nil {
		log.Printf("error creating table log: %s", err)
		return
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `tags` ( `user` BIGINT NOT NULL, `tag` VARCHAR(20) NOT NULL, PRIMARY KEY (`user`, `tag`))")
	if err != nil {
		log.Printf("error creating table memberships: %s", err)
	}

	return err
}

func ensureDB() {
	databaseOnce.Do(func() {
		err := openDB()
		if err != nil {
			log.Fatalf("Tried to ensure DB: %s", err)
		}
	})
}

func StartDB() error {
	var err error

	databaseOnce.Do(func() {
		err = openDB()
	})

	return err
}

var (
	stmtMap   = make(map[string]*sql.Stmt)
	stmtMapMu sync.RWMutex
)

func StmtExec(query string, args ...any) (r sql.Result, err error) {
	stmtMapMu.Lock()
	defer stmtMapMu.Unlock()

	ensureDB()

	stmt, ok := stmtMap[query]
	if !ok {
		stmt, err = database.Prepare(query)
		if err != nil {
			return
		}

		stmtMap[query] = stmt
	}

	return stmt.Exec(args...)
}

func StmtQuery(query string, args ...any) (r *sql.Rows, err error) {
	stmtMapMu.Lock()
	defer stmtMapMu.Unlock()

	ensureDB()

	stmt, err := database.Prepare(query)
	if err != nil {
		return
	}

	defer stmt.Close()

	return stmt.Query(args...)
}

func UserHasTag(user int64, tag string) (bool, error) {
	r, err := StmtQuery("SELECT 1 FROM tags WHERE (user = ? AND tag = ?)",
		user, tag,
	)
	if err != nil {
		return false, err
	}

	defer r.Close()

	return r.Next(), nil
}

func SetUserTag(user int64, tag string) error {
	_, err := StmtExec("INSERT IGNORE INTO tags (user, tag) VALUES (?, ?);",
		user, tag,
	)

	return err
}

func SetGroupTag(group int64, tag string) error {
	_, err := StmtExec("INSERT INTO group_tags (group_id, tag) VALUES (?, ?);",
		group, tag,
	)

	return err
}

// on len(0) returns nil []
func GetTaggedGroups(tag string) (groups []int64, err error) {
	r, err := StmtQuery("SELECT group_id FROM group_tags WHERE tag = ?;",
		tag,
	)

	var buf int64
	for r.Next() {
		err = r.Scan(&buf)
		if err != nil {
			return
		}

		groups = append(groups, buf)
	}

	return
}

// removes tag from user, bool is true if tag was present, false if it was not set
func RmUserTag(user int64, tag string) (bool, error) {
	r, err := StmtExec("DELETE FROM tags WHERE user = ? AND tag = ?;",
		user, tag,
	)
	if err != nil {
		return false, err
	}

	changed, err := r.RowsAffected()

	return changed > 0, err
}

// removes given tag from all users, bool is true if tag was present at all, false if it was not set anywhere
func RmAllUserTag(tag string) (bool, error) {
	r, err := StmtExec("DELETE FROM tags WHERE tag = ?;",
		tag,
	)
	if err != nil {
		return false, err
	}

	changed, err := r.RowsAffected()

	return changed > 0, err
}

func GetTaggedUsers() (s []int64, err error) {
	r, err := StmtQuery("SELECT user FROM tags")
	if err != nil {
		return nil, err
	}

	var buf int64
	for r.Next() {
		if err = r.Scan(&buf); err != nil {
			return
		}

		s = append(s, buf)
	}

	return
}

func GetUsersWithTag(tag string) (s []int64, err error) {
	r, err := StmtQuery("SELECT user FROM tags WHERE (tag = ?);",
		tag,
	)
	if err != nil {
		return nil, err
	}

	var buf int64
	for r.Next() {
		if err = r.Scan(&buf); err != nil {
			return
		}

		s = append(s, buf)
	}

	return
}

func GetUserTags(user int64) (s []string, err error) {
	r, err := StmtQuery("SELECT tag FROM tags WHERE (user = ?);",
		user,
	)
	if err != nil {
		return nil, err
	}

	var buf string
	for r.Next() {
		if err = r.Scan(&buf); err != nil {
			return
		}

		s = append(s, buf)
	}

	return
}

func SetArrival(user int64, time int64) error {
	_, err := StmtExec("INSERT INTO arrivalTimes (user, time) VALUES (?, ?) ON DUPLICATE KEY UPDATE time = VALUES(time);",
		user, time,
	)

	if err != nil {
		log.Printf("Error Setting Arrival: %s", err)
	}

	return err
}

// moves users arrival time by time in s
func MoveArrival(user int64, time int64) (bool, error) {
	r, err := StmtExec("UPDATE arrivalTimes SET time = time + ? WHERE user = ?",
		time, user,
	)

	if err != nil {
		log.Printf("Error Moving Arrival Time: %s", err)

		return false, err
	}

	i, err := r.RowsAffected()
	return i > 0, err
}

func RmArrival(user int64) (bool, error) {
	r, err := StmtExec("DELETE FROM arrivalTimes WHERE user = ?;",
		user,
	)

	if err != nil {
		log.Printf("Error Removing Arrival: %s", err)

		return false, err
	}

	i, err := r.RowsAffected()
	return i > 0, err
}

type Arrival struct {
	User int64
	Time int64
}

func GetArrivals() ([]Arrival, error) {
	DoCleanArrivals()

	r, err := StmtQuery("SELECT user, time FROM arrivalTimes")
	if err != nil {
		return nil, err
	}

	var user, time int64
	a := make([]Arrival, 0)

	for r.Next() {
		err = r.Scan(&user, &time)
		if err != nil {
			return nil, err
		}

		a = append(a, Arrival{user, time})
	}

	return a, err
}

func SetLocation(user int64, since int64, note string) error {
	_, err := StmtExec("INSERT INTO location (user, since, note) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE note = VALUES(note), since = VALUES(since);",
		user, since, note,
	)

	if err != nil {
		log.Printf("Error Setting Location: %s", err)
	}

	return err
}

func SetLocationDepart(user int64) (changed bool, err error) {
	r, err := StmtExec("DELETE FROM location WHERE user = ?;",
		user,
	)

	if err != nil {
		return false, err
	}

	i, err := r.RowsAffected()
	return i > 0, err
}

func DoCleanArrivals() {
	rows, err := CleanArrivals()
	if err != nil {
		log.Printf("Failed to clean! %s", err)
		return
	} else {
		if rows > 0 {
			log.Printf("Cleaned %d rows", rows)
		}
	}
}

// removes all entries from !today
// returns rows affected
func CleanArrivals() (int64, error) {
	morning := util.Today(0).Unix()
	evening := util.Today((time.Hour * 24) - 1).Unix()

	r, err := StmtExec(`DELETE FROM arrivalTimes
	WHERE time < ? OR time > ?;`, morning, evening)

	if err != nil {
		return 0, err
	}

	rows, err := r.RowsAffected()
	return rows, err
}

type UserState struct {
	ID    int64
	Since int64

	Note string
}

func IsUserThere(u int64) (there bool, note string, err error) {
	DoCleanArrivals()

	r, err := StmtQuery("SELECT 1, note FROM location WHERE user = ?",
		u)
	if err != nil {
		return false, "", err
	}

	if r.Next() {
		err = r.Scan(&there, &note)

		return
	}

	return false, "", nil
}

func WhoThere() ([]UserState, error) {
	DoCleanArrivals()

	r, err := StmtQuery("SELECT user, since, note FROM location")
	if err != nil {
		return nil, err
	}

	i := make([]UserState, 0)
	var buf UserState
	for r.Next() {
		err := r.Scan(&buf.ID, &buf.Since, &buf.Note)
		if err != nil {
			return nil, err
		}

		i = append(i, buf)
	}

	return i, err
}

func CleanLocation() (int64, error) {
	r, err := StmtExec("DELETE FROM location WHERE since > ?",
		util.Today((-24+4)*time.Hour).Unix(),
	)
	if err != nil {
		return 0, err
	}

	return r.RowsAffected()
}

// `user` BIGINT NOT NULL, `time` BIGINT NOT NULL, `key` VARCHAR(20) NOT NULL, `value` TEXT NOT NULL, PRIMARY KEY (`key`, `time`, `key`))")
func AddLog(user int64, time time.Time, key string, value string) error {
	_, err := StmtExec("INSERT INTO `log` (user, time, `key`, value) VALUES (?, ?, ?, ?)",
		user, time.Unix(), key, value,
	)

	return err
}

type Log struct {
	User  int64
	Time  int64 // unixsecs
	Key   string
	Value string
}

func QueryLog(key string) ([]Log, error) {
	r, err := StmtQuery("SELECT user, time, key, values FROM log WHERE key = ?", key)
	if err != nil {
		return nil, err
	}

	var l []Log
	var buf Log

	for r.Next() {
		err = r.Scan(&buf.User, &buf.Time, &buf.Key, &buf.Value)
		if err != nil {
			return nil, err
		}

		l = append(l, buf)
	}

	return l, nil
}
