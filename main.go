package main

import (
	"log"
	"os"

	"github.com/gocql/gocql"
	"github.com/jochenboesmans/forward-financial-statements/predict"
	"github.com/jochenboesmans/forward-financial-statements/tickers"
	"github.com/joho/godotenv"
)

var dbSession *gocql.Session

func main() {
	switch os.Args[1] {
	case "pull":
	case "predict":
		predict.Predict()
	case "write-tickers-to-db":
		err := tickers.WriteFileContentToDb(dbSession)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println("done")
	case "read-tickers-from-db":
		ts, err := tickers.ReadTickersFromDb(dbSession)
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(ts)
	}
}

func init() {
	godotenv.Load()
	dbSetup()
}

func dbSetup() {
	cluster := gocql.NewCluster(os.Getenv("CASSANDRA_IP"))
	cluster.ProtoVersion = 4
	cluster.Keyspace = os.Getenv("CASSANDRA_KEYSPACE")
	newDbSession, err := cluster.CreateSession()
	if err != nil {
		log.Fatalf("could not connect to cassandra cluster: %v", err)
	}

	dbSession = newDbSession
}
