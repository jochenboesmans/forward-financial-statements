package pull

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

type IncomeStatementTimeSeries []IncomeStatement
type IncomeStatement struct {
	Revenue   float64 `json:"revenue"`
	NetIncome float64 `json:"netIncome"`
}

func getIncomeStatements(ticker string, apiKey string) IncomeStatementTimeSeries {
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

func (ists IncomeStatementTimeSeries) Revenues() []float64 {
	r := []float64{}
	for _, v := range ists {
		r = append(r, v.Revenue)
	}
	return r
}

func (ists IncomeStatementTimeSeries) NetIncomes() []float64 {
	r := []float64{}
	for _, v := range ists {
		r = append(r, v.NetIncome)
	}
	return r
}

func Pull() {
	apiKey := os.Getenv("API_KEY")
	tickers := getTickers()

	incomeStatements := map[string]IncomeStatementTimeSeries{}
	for _, t := range tickers {
		incomeStatements[t] = getIncomeStatements(t, apiKey)
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
