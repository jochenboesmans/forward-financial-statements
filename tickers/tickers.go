package tickers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gocql/gocql"
)

type tickers []string

func WriteFileContentToDb(dbSession *gocql.Session) error {
	ts, err := readTickersFromDisk()
	if err != nil {
		return err
	}
	ts.writeToDb(dbSession)
	return nil
}

func readTickersFromDisk() (tickers, error) {
	file, err := ioutil.ReadFile("tickers.json")
	if err != nil {
		return nil, err
	}
	t := tickers{}
	err = json.Unmarshal(file, &t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (ts tickers) writeToDb(dbSession *gocql.Session) {
	keySpaceMeta, _ := dbSession.KeyspaceMetadata(os.Getenv("CASSANDRA_KEYSPACE"))

	if _, exists := keySpaceMeta.Tables["tickers"]; exists != true {
		fmt.Println("not exists")
		dbSession.Query("CREATE TABLE tickers (" +
			"id text, " +
			"PRIMARY KEY (id))").Exec()
	} else {
		dbSession.Query("DELETE FROM tickers").Exec()
	}

	for _, t := range ts {
		dbSession.Query("INSERT INTO tickers (id) VALUES (?)", t).Exec()
	}
}

func ReadTickersFromDb(dbSession *gocql.Session) (tickers, error) {
	keySpaceMeta, _ := dbSession.KeyspaceMetadata(os.Getenv("CASSANDRA_KEYSPACE"))

	if _, exists := keySpaceMeta.Tables["tickers"]; exists != true {
		return tickers{}, errors.New("tickers table doesn't exist")
	}

	ts := tickers{}
	t := ""
	iter := dbSession.Query("SELECT id FROM tickers").Iter()
	for iter.Scan(&t) {
		ts = append(ts, t)
	}
	return ts, nil
}
