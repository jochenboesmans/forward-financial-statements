package market_cap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type marketCapResponse []marketCap
type marketCap struct {
	MarketCap float64 `json:"marketCap"`
}

func GetMarketCap(ticker string) float64 {
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

	r := marketCapResponse{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		fmt.Println(err)
	}

	if len(r) > 0 {
		return r[0].MarketCap
	}
	return float64(0)
}
