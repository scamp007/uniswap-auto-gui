package services

import (
	"encoding/json"
	"sync"

	"github.com/uniswap-auto-gui/utils"
)

func StableTokens(wg *sync.WaitGroup, pairs utils.Pairs) {
	defer wg.Done()

	for _, item := range pairs.Data.Pairs {
		c := make(chan string, 1)
		go utils.Post(c, "swaps", item.Id)
		stableToken(c, item.Id)
	}
}

func stableToken(pings <-chan string, id string) {
	var swaps utils.Swaps
	msg := <-pings
	json.Unmarshal([]byte(msg), &swaps)

	if len(swaps.Data.Swaps) > 0 {
		min, max, _, _, _, _ := minMax(swaps)
		last, _ := priceOfSwap(swaps.Data.Swaps[0])
		_, _, period := periodOfSwaps(swaps)

		if (max-min)/last > 0.1 && period < 6 {

		} else if (max-min)/last < 0.1 && period > 24 {

		}
	}
}
