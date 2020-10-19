package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/joho/godotenv"
	"github.com/sajari/regression"
)

type IncomeStatementTimeSeries []IncomeStatement
type IncomeStatement struct {
	Revenue   float64 `json:"revenue"`
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
	switch os.Args[1] {
	case "pull":
		pull()
	case "predict":
		predict()
	}
}

func init() {
	godotenv.Load()
}

func (prwt PredictionResultsWithTicker) format() string {
	r := ""
	for _, v := range prwt {
		r += fmt.Sprintf("%s:\n", v.Ticker)
		r += fmt.Sprintf("forward quarterly P/E: %+v\n", v.PredictionResult.PES)
		r += fmt.Sprintf("forward quarterly P/R: %+v\n", v.PredictionResult.PRS)
	}
	return r
}
func (prwt PredictionResultsWithTicker) Len() int      { return len(prwt) }
func (prwt PredictionResultsWithTicker) Swap(i, j int) { prwt[i], prwt[j] = prwt[j], prwt[i] }
func (prwt PredictionResultsWithTicker) Less(i, j int) bool {
	return len(prwt[i].PRS) != 0 && len(prwt[j].PRS) != 0 && prwt[i].PRS[len(prwt[i].PRS)-1] < prwt[j].PRS[len(prwt[j].PRS)-1]
}

type PredictionResultsWithTicker []PredictionResultWithTicker

func (pr *PredictionResults) sort() PredictionResultsWithTicker {
	array := []PredictionResultWithTicker{}
	for k, v := range *pr {
		array = append(array, PredictionResultWithTicker{
			PredictionResult: v,
			Ticker:           k,
		})
	}

	sort.Slice(array, func(i, j int) bool {
		return len(array[i].PRS) != 0 && len(array[j].PRS) != 0 && array[i].PRS[len(array[i].PRS)-1] > array[j].PRS[len(array[j].PRS)-1]
	})
	return array
}

type PredictionResults map[string]PredictionResult
type PredictionResult struct {
	PES []float64
	PRS []float64
}
type PredictionResultWithTicker struct {
	PredictionResult
	Ticker string
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

	predictionResults := PredictionResults{}
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
		for i := range []int{0, 1, 2, 3, 4} {
			prediction, err := r.Predict([]float64{float64(len(revenues) + i)})
			if err == nil {
				predictions = append(predictions, prediction)
			}
		}

		netIncomes := ists.netIncomes()

		r = new(regression.Regression)
		r.SetObserved("NetIncome")
		r.SetVar(0, "Quarter")
		for i, netInc := range netIncomes {
			train(netInc, i, r)
		}
		r.Run()

		predictionsNetIncomes := []float64{}
		for i := range []int{0, 1, 2, 3, 4} {
			prediction, err := r.Predict([]float64{float64(len(netIncomes) + i)})
			if err == nil {
				predictionsNetIncomes = append(predictionsNetIncomes, prediction)
			}
		}

		mcap := getMarketCap(ticker)
		pes := []float64{}
		for _, pni := range predictionsNetIncomes {
			pes = append(pes, float64(mcap/(pni*4)))
		}
		prs := []float64{}
		for _, p := range predictions {
			prs = append(prs, float64(mcap/(p*4)))
		}
		predictionResult := PredictionResult{
			PES: pes,
			PRS: prs,
		}
		predictionResults[ticker] = predictionResult
	}

	sorted := predictionResults.sort()

	forwardValuations := sorted.format()

	err = ioutil.WriteFile("cloudflare-fw-valuation.txt", []byte(forwardValuations), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("all done")
}

type MarketCapResponse []MarketCap
type MarketCap struct {
	MarketCap float64 `json:"marketCap"`
}

func getMarketCap(ticker string) float64 {
	apiKey := os.Getenv("API_KEY")
	url := fmt.Sprintf("https://financialmodelingprep.com/api/v3/market-capitalization/%s?apikey=%s", ticker, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		fmt.Println(err)
	}

	r := MarketCapResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println(err)
	}

	if len(r) > 0 {
		return r[0].MarketCap
	}
	return float64(0)
}

func train(revenue float64, i int, r *regression.Regression) {
	dp := regression.DataPoint(revenue, []float64{float64(i)})
	r.Train(dp)
}

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

func pull() {
	apiKey := os.Getenv("API_KEY")
	tickers := getTickers()

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
