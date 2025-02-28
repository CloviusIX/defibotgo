package services

import (
	"defibotgo/internal/models"
	"defibotgo/internal/utils"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
)

var dexscreenerUrl = "https://api.dexscreener.com/latest/dex/pairs"
var opWethVelo = "0x58e6433A6903886E440Ddf519eCC573c4046a6b2"
var opWethAero = "0x7f670f78B17dEC44d5Ef68a48740b6f8849cc2e6"

type Pair struct {
	PriceNative string `json:"priceNative"`
}

type DexScreenerResponse struct {
	Pairs []Pair `json:"pairs"`
}

func getPairAddress(chain models.Chain) string {
	switch chain {
	case models.Base:
		return opWethAero
	default:
		return opWethVelo
	}
}

func GetPoolPrice(chain models.Chain) (*big.Int, error) {
	pairAddress := getPairAddress(chain)
	url := fmt.Sprintf("%s/%s/%s", dexscreenerUrl, strings.ToLower(string(chain)), pairAddress)
	response, err := get(url)
	if err != nil {
		return nil, fmt.Errorf("fail to get response %v", err)
	}

	var dexScreenerResponse DexScreenerResponse
	err = json.Unmarshal(response, &dexScreenerResponse)
	if err != nil {
		return nil, fmt.Errorf("fail to unmarshal response %v", err)
	}

	if len(dexScreenerResponse.Pairs) == 0 {
		return nil, fmt.Errorf("no pairs found in response")
	}

	priceNativeBigInt, err := utils.ParseWeiString(dexScreenerResponse.Pairs[0].PriceNative)

	if err != nil {
		return nil, fmt.Errorf("priceNative could not be parsed: %v", err)
	}

	return priceNativeBigInt, nil
}

func get(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("could not fetch api: %v", err)
	}

	defer response.Body.Close()

	// Check for non-200 status code
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	return body, nil
}
