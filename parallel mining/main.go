package main

import (
	"crypto/sha256"
	"fmt"
	"math"
	"math/rand"
	"sync"
)

type Block struct {
	ID   int
	Data string
}

type Result struct {
	MinerID int
	BlockID int
	Nonce   int
	Hash    string
}

func mineBlock(block Block, minerCount int, difficulty int, score map[int]int) Result {
	done := make(chan struct{})    // closed the instant someone wins
	winner := make(chan Result, 1) // buffered so the winning send never blocks
	var once sync.Once
	var wg sync.WaitGroup

	for m := 0; m < minerCount; m++ {
		wg.Add(1)
		go func(minerID int) {
			defer wg.Done()

			for {
				select {
				case <-done:
					return
				default:
					nonce := rand.Intn(math.MaxInt32)
					hash := sha256.Sum256([]byte(block.Data + fmt.Sprint(nonce)))
					if hasLeadingZeroBits(hash[:], difficulty) {
						once.Do(func() {
							winner <- Result{MinerID: m, BlockID: block.ID, Nonce: nonce, Hash: fmt.Sprint(hash)}
							close(done)
						})
					}
				}
			}
		}(m)
	}

	wg.Wait() // block until every miner has stopped (won, or bailed via done)
	result := <-winner
	score[result.MinerID]++
	return result
}

func hasLeadingZeroBits(hash []byte, n int) bool {
	if n < 0 || n > len(hash)*8 {
		return false
	}

	fullBytes := n / 8
	remainingBits := n % 8

	for i := 0; i < fullBytes; i++ {
		if hash[i] != 0 {
			return false
		}
	}
	if remainingBits > 0 {
		mask := byte(0xFF << (8 - remainingBits))
		if hash[fullBytes]&mask != 0 {
			return false
		}
	}
	return true
}

func main() {
	const numMiners = 5
	const difficulty = 22 // tune so each block takes a moment but not forever
	blocks := []Block{
		{ID: 0, Data: "block-0"},
		{ID: 1, Data: "block-1"},
		{ID: 2, Data: "block-2"},
		{ID: 3, Data: "block-3"},
		{ID: 4, Data: "block-4"},
	}

	score := make(map[int]int)
	var ledger []Result

	for _, b := range blocks {
		result := mineBlock(b, numMiners, difficulty, score)
		ledger = append(ledger, result)
		fmt.Printf("block %d won by miner %d (nonce=%d)\n", result.BlockID, result.MinerID, result.Nonce)
	}

	fmt.Println("final scores:", score)
	fmt.Println("blocks mined:", len(ledger))
}
