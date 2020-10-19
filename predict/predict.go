package predict

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/jochenboesmans/forward-financial-statements/market_cap"
	"github.com/jochenboesmans/forward-financial-statements/pull"
	"github.com/sajari/regression"
)

type PredictionResultsWithTicker []PredictionResultWithTicker

func (prwt PredictionResultsWithTicker) Len() int      { return len(prwt) }
func (prwt PredictionResultsWithTicker) Swap(i, j int) { prwt[i], prwt[j] = prwt[j], prwt[i] }
func (prwt PredictionResultsWithTicker) Less(i, j int) bool {
	return len(prwt[i].PRS) != 0 && len(prwt[j].PRS) != 0 && prwt[i].PRS[len(prwt[i].PRS)-1] < prwt[j].PRS[len(prwt[j].PRS)-1]
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

func Predict() {
	data, err := ioutil.ReadFile("financial-statements.json")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	istss := map[string]pull.IncomeStatementTimeSeries{}
	err = json.Unmarshal(data, &istss)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	predictionResults := PredictionResults{}
	for ticker, ists := range istss {
		revenues := ists.Revenues()

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

		netIncomes := ists.NetIncomes()

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

		mcap := market_cap.GetMarketCap(ticker)
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

	err = ioutil.WriteFile("forward-valuations.txt", []byte(forwardValuations), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("all done")
}

func train(revenue float64, i int, r *regression.Regression) {
	dp := regression.DataPoint(revenue, []float64{float64(i)})
	r.Train(dp)
}
