package tickers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type tickers []string

func getTickers() tickers {
	file, err := ioutil.ReadFile("tickers.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t := tickers{}
	err = json.Unmarshal(file, &t)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return t
}
