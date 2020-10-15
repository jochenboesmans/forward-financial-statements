package main

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"os"
	"github.com/joho/godotenv"
	"github.com/sajari/regression"
)

type IncomeStatementTimeSeries []IncomeStatement
type IncomeStatement struct {
	Revenue float64 `json:"revenue"`
	NetIncome float64 `json:"netIncome"`
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

func (ists IncomeStatementTimeSeries) revenues() []float64 {
	r := []float64{}
	for _, v := range ists {
		r = append(r, v.Revenue)
	}
	return r
}

func (ists IncomeStatementTimeSeries) netIncomes() []float64 {
	r := []float64{}
	for _, v := range ists {
		r = append(r, v.NetIncome)
	}
	return r
}

func main() {
	switch (os.Args[1]) {
		case "pull": pull()
		case "predict": predict()
	}
}

func init() {
	godotenv.Load()
}

func predict() {
	data, err := ioutil.ReadFile("financial-statements.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	istss := map[string]IncomeStatementTimeSeries{}
	err = json.Unmarshal(data, &istss)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//predictedIstss := map[string]IncomeStatementTimeSeries{}
	for ticker, ists := range istss {
		revenues := ists.revenues()

		r := new(regression.Regression)
		r.SetObserved("Revenue")
		r.SetVar(0, "Quarter")
		for i, rev := range revenues {
			train(rev, i, r)
		}
		r.Run()

		predictions := []float64{}
		for i := range []int{0,1,2,3,4} {
			prediction, err := r.Predict([]float64{float64(len(revenues)+i)})
			if err == nil {
				predictions = append(predictions, prediction)
			}
		}
		fmt.Printf("%s: %+v\n", ticker, predictions)
	}

}

func train(revenue float64, i int, r *regression.Regression) {
	dp := regression.DataPoint(revenue, []float64{float64(i)})
	r.Train(dp)
}

func pull() {
	apiKey := os.Getenv("API_KEY")
	tickers := []string{"NFLX", "SPOT", "TSLA", "LYFT", "TWTR", "FB", "MA", "UBER", "DAL", "AMZN", "DELL", "V", "SHOP", "MSFT", "AAPL", "NVDA", "AMD", "SQ", "INTC", "LEVI", "MU", "GOOG", "WORK", "DIS", "DOCU", "IBKR", "TKWY", "SPCE", "GPRO", "PTON", "ZM"}

	incomeStatements := map[string]IncomeStatementTimeSeries{}
	for _, t := range tickers {
		incomeStatements[t] = request(t, apiKey)
	}

	incomeStatementsJSON, err := json.Marshal(incomeStatements)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = ioutil.WriteFile("financial-statements.json", incomeStatementsJSON, 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("all done")
}
