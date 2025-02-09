package main_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"algorillas.com/monko/events"
	"algorillas.com/monko/internal"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	monkoID       = uint64(2494786278)
	watcher       = internal.NewWatcher()
	blockAnalyzer = internal.NewBlockAnalyzer(monkoID)
)

func TestLiquidityEvent(t *testing.T) {

	round := uint64(47026784)

	info, err := GetBlockForRound(round)
	require.Nil(t, err)

	reports := blockAnalyzer.AnalyzeBlock(info)

	foundEvents := []events.Event{}
	for _, report := range reports {
		event := watcher.Event(report)

		bytes, err := json.MarshalIndent(event, "", "  ")
		require.Nil(t, err)

		fmt.Printf("Event: %s\n", string(bytes))

		foundEvents = append(foundEvents, event)

	}

	assert.Equal(t, 1, len(foundEvents))

	event := foundEvents[0]

	assert.Equal(t, events.AddAction, event.EventAction())
}

func GetBlockForRound(round uint64) (types.Block, error) {

	client, err := algod.MakeClient("https://mainnet-api.4160.nodely.dev", "")
	if err != nil {
		return types.Block{}, err
	}

	info, err := client.Block(round).Do(context.Background())
	if err != nil {
		return types.Block{}, err
	}

	return info, nil
}
