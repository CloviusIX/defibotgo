package asyncservices

import (
	"defibotgo/internal/models"
	"defibotgo/internal/services"
	"defibotgo/internal/utils"
	"github.com/dgraph-io/ristretto"
	"math/big"
	"sync"
)

func GetPoolPriceAsync(chain models.Chain, cache *ristretto.Cache, cacheKey string, ch chan models.WeiResult, wg *sync.WaitGroup) {
	defer wg.Done()
	if cacheResult, found := cache.Get(cacheKey); found {
		ch <- models.WeiResult{Value: cacheResult.(*big.Int), Err: nil}
		return
	}
	pairPrice, err := services.GetPoolPrice(chain)
	if err != nil {
		ch <- models.WeiResult{Value: big.NewInt(0), Err: err}
	}

	cache.SetWithTTL(cacheKey, pairPrice, 1, utils.CacheTime)
	ch <- models.WeiResult{Value: pairPrice, Err: nil}

}
