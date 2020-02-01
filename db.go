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

func AddTopicData(topic_name string, value float64) error {
	err := dbHandlers[dbTableKind("TOPIC_DATA_"+topic_name)].Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(strconv.Itoa(int(time.Now().Unix()))), []byte(strconv.FormatFloat(value, 'f', -1, 64)))
		return err
	})
	return err
}

func prepareDirectory(dir ...string) {
	for _, val := range dir {
		if _, err := os.Stat(val); os.IsNotExist(err) {
			os.Mkdir(val, 0666)
		}
	}
}
