package main

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/json-iterator/go"
)

const (
	UnconfirmedTXsNum = "http://localhost:26657/num_unconfirmed_txs"
	PostTx            = "http://localhost:26657/broadcast_tx_async?tx="
	RPS         	  = 3000

	logStep = 1000
)

var TxTime = time.Duration(big.NewInt(0).Div(big.NewInt(int64(time.Second)), big.NewInt(RPS)).Int64()) // ns

func main() {
	done := new(uint32)

	hourTimer := time.NewTimer(time.Hour)
	defer hourTimer.Stop()

	round := time.NewTicker(TxTime)
	defer round.Stop()

	i := 0
	var totalTime time.Duration
	mainTime := time.Now()

mainLoop:
	for {
		startTime := time.Now()

		select {
		case <-hourTimer.C:
			break mainLoop
		case <-round.C:
			postTxs(i, i+1, done)
		}

		endTime := time.Now()

		roundTime := endTime.Sub(startTime)
		totalTime += roundTime
		currentDuration := endTime.Sub(mainTime)

		fmt.Printf("Total time for round %v: %v. Total test duration %v. RPS: %v\n", i, roundTime, currentDuration, float64(i+1)/float64(currentDuration)*float64(time.Second))

		i++

		if i % 1000 == 0 {
			hasUnconfirmedTxs(true)
		}
	}

	// wait until all txs passed
	time.Sleep(50 * time.Millisecond)
	for !hasUnconfirmedTxs(false) {
		time.Sleep(50 * time.Millisecond)
	}

	fmt.Println("Done", i)
	fmt.Println("Total time", totalTime)
}

func postTxs(from, to int, done *uint32) {
	for i := from; i < to; i++ {
		postTx(i)
		done := atomic.AddUint32(done, 1)

		if done%logStep == 0 {
			fmt.Println("Already done", done)
		}
	}
}

func postTx(n int) {
	doRequest(PostTx + "\"" + strconv.Itoa(time.Now().Nanosecond()) + strconv.Itoa(n) + "\"")
}

func hasUnconfirmedTxs(withLog bool) bool {
	res := doRequest(UnconfirmedTXsNum)

	resp := new(RPCResponse)
	resp.Decode(res)

	if withLog {
		fmt.Println("Has Unconfirmed Txs", string(res))
	}

	n, err := strconv.Atoi(resp.Res.N)
	if err != nil {
		fmt.Printf("error while getting unconfirmed TXs: %v, %q\n", err, string(res))
		return true
	}
	return n == 0
}

func doRequest(url string) []byte {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("error while http.get", err)
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error while reading response body", err)
		return nil
	}

	return body
}

type RPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      string `json:"id"`
	Res     Result `json:"result"`
}

type Result struct {
	N   string `json:"n_txs"`
	Txs *uint  `json:"txs"`
}

func (r *RPCResponse) Decode(input []byte) {
	var json = jsoniter.ConfigFastest
	json.Unmarshal(input, r)
}
