package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"github.com/joho/godotenv"
)

type IncomeStatementTimeSeries []IncomeStatement
type IncomeStatement struct {
	Revenue int `json:"revenue"`
	NetIncome int `json:"netIncome"`
}

func request(ticker string, apiKey string) IncomeStatementTimeSeries {
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/income-statement/%s?period=quarter&limit=400&apikey=%s", ticker, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return IncomeStatementTimeSeries{}
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println(err)
		return IncomeStatementTimeSeries{}
	}

	r := IncomeStatementTimeSeries{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println(err)
		return IncomeStatementTimeSeries{}
	}
	r.reverse()
	return r
}

func (ists IncomeStatementTimeSeries) reverse() {
	r := IncomeStatementTimeSeries{}
	for i := range ists {
		r = append(r, ists[len(ists)-1-i])
	}

	for i := range r {
		ists[i] = r[i]
	}
}

func main() {
	apiKey := os.Getenv("API_KEY")
	tickers := []string{"NFLX", "SPOT"}

	incomeStatements := map[string]IncomeStatementTimeSeries{}
	for _, t := range tickers {
		incomeStatements[t] = request(t, apiKey)
	}
	fmt.Println(incomeStatements)
}

func init() {
	godotenv.Load()
}
