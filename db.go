package main

import (
	badger "github.com/dgraph-io/badger"
	"log"
	"os"
	"strconv"
	"time"
)

type dbTableKind string

var dbHandlers map[dbTableKind]*badger.DB

func OpenDB(root_dir string) {
	dbHandlers = make(map[dbTableKind]*badger.DB)
	for name, _ := range topicDetail {
		dbname := dbTableKind("TOPIC_DATA_" + name)
		dir := root_dir + "/db/" + string(dbname)
		prepareDirectory(dir)
		db, err := badger.Open(badger.DefaultOptions(dir).WithTruncate(true))
		if err != nil {
			log.Fatal(err)
		} else {
			dbHandlers[dbname] = db
		}
	}
}

func CloseDB() {
	for _, hnd := range dbHandlers {
		hnd.Close()
	}
}

func AddValue(topic_name string, data *topicData) error {
	err := getDbHandler(topic_name).Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(strconv.Itoa(data.Time)), []byte(strconv.FormatFloat(data.Value, 'f', -1, 64)))
		return err
	})
	return err
}

func GetValue(topic_name string, term string, max_count int) ([]*topicData, error) {
	//term: 1s, 1m, 1h, 1d, 1m
	var last_time_key, now_count int
	last_time_key = 99999999999999
	topic_by_term := []*topicData{}

	err := getDbHandler(topic_name).View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Reverse = true
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
				if isAnotherTerm(pk, last_time_key, term) {
					now_count++
					topic_by_term = append(topic_by_term, &topicData{pk, pv})
					last_time_key = pk
				}
				return nil
			})
			if err != nil {
				return err
			}
			if now_count >= max_count {
				return nil
			}
		}
		return nil
	})
	return topic_by_term, err
}

func isAnotherTerm(last_time int, now_time int, term string) bool {
	clast_time := time.Unix(int64(last_time), 0)
	cnow_time := time.Unix(int64(now_time), 0)
	diff := cnow_time.Sub(clast_time)

	var param, difflimit int
	param, _ = strconv.Atoi(string(term[0 : len(term)-1]))

	switch string(term[len(term)-1]) {
	case "s":
		difflimit = 1
	case "m":
		difflimit = 60
	case "h":
		difflimit = 3600
	case "d":
		difflimit = 86400
	}

	if diff.Seconds() >= float64(param*difflimit) {
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
