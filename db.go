package main

import (
	badger "github.com/dgraph-io/badger"
	"log"
	"os"
	"strconv"
	"time"
)

type dbTableKind string

var dbTableNames = [...]dbTableKind{"TOPIC_DATA_test", "TOPIC_DATA_btcusd"}
var dbHandlers map[dbTableKind]*badger.DB

func OpenDB(root_dir string) {
	dbHandlers = make(map[dbTableKind]*badger.DB)
	for _, name := range dbTableNames {
		dir := root_dir + "/db/" + string(name)
		prepareDirectory(dir)
		db, err := badger.Open(badger.DefaultOptions(dir))
		if err != nil {
			log.Fatal(err)
		} else {
			dbHandlers[name] = db
		}
	}
}

func CloseDB() {
	for _, hnd := range dbHandlers {
		hnd.Close()
	}
}

func AddTopicData(topic_name string, data *topicData) error {
	err := getDbHandler(topic_name).Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(strconv.Itoa(data.Time)), []byte(strconv.FormatFloat(data.Value, 'f', -1, 64)))
		return err
	})
	return err
}

func GetTopicData(topic_name string, term string) ([]*topicData, error) {
	//term: 1s, 1m, 1h, 1d, 1m
	var last_time_key int
	topic_by_term := []*topicData{}

	err := getDbHandler(topic_name).View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			k := item.Key()
			err := item.Value(func(v []byte) error {
				pk, err := strconv.Atoi(string(k))
				if err != nil {
					return err
				}
				pv, err := strconv.ParseFloat(string(v), 64)
				if err != nil {
					return err
				}

				if isAnotherTerm(last_time_key, pk, term) {
					topic_by_term = append(topic_by_term, &topicData{pk, pv})
				}
				last_time_key = pk
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return topic_by_term, err
}

func isAnotherTerm(last_time int, now_time int, term string) bool {
	clast_time := time.Unix(int64(last_time), 0)
	cnow_time := time.Unix(int64(now_time), 0)
	var before, after int
	switch term {
	case "1s":
		before = clast_time.Second()
		after = cnow_time.Second()
	case "1m":
		before = clast_time.Minute()
		after = cnow_time.Minute()
	case "1h":
		before = clast_time.Hour()
		after = cnow_time.Hour()
	}
	if before < after {
		return true
	} else {
		return false
	}
}

func getDbHandler(topic_name string) *badger.DB {
	val, ok := dbHandlers[dbTableKind("TOPIC_DATA_"+topic_name)]
	if !ok {
		log.Panicf("No Db handler key matched with : %s", "TOPIC_DATA_"+topic_name)
	}
	return val
}

func prepareDirectory(dir ...string) {
	for _, val := range dir {
		if _, err := os.Stat(val); os.IsNotExist(err) {
			os.Mkdir(val, 0666)
		}
	}
}
