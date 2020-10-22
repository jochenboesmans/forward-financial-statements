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
	err = ts.writeToDb(dbSession)
	if err != nil {
		return err
	}
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

func (ts tickers) writeToDb(dbSession *gocql.Session) error {
	keySpaceMeta, _ := dbSession.KeyspaceMetadata(os.Getenv("CASSANDRA_KEYSPACE"))

	if _, exists := keySpaceMeta.Tables["tickers"]; exists != true {
		// create table
		dbSession.Query("CREATE TABLE tickers (" +
			"id text, " +
			"PRIMARY KEY (id));").Exec()
	} else {
		// wipe existing table
		dbSession.Query("TRUNCATE tickers").Exec()
	}

	batch := dbSession.NewBatch(gocql.LoggedBatch)
	for _, t := range ts {
		batch.Query("INSERT INTO tickers (id) VALUES (?);", t)
	}

	err := dbSession.ExecuteBatch(batch)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
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
