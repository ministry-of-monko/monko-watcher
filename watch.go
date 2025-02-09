package main

import (
	"context"
	"fmt"

	"algorillas.com/monko/events"
	"algorillas.com/monko/internal"
)

func main() {

	watcher := internal.NewWatcher()
	blockAnalyzer := internal.NewBlockAnalyzer(watcher.Config.Asset.ID)

	status, err := watcher.AlgodClient.Status().Do(context.Background())
	if err != nil {
		panic(err)
	}

	start := status.LastRound
	round := status.LastRound

	for {

		watcher.CalcPrices(start, round)
		watcher.UpdateHolderCount(start, round)

		_, err := watcher.AlgodClient.StatusAfterBlock(round).Do(context.Background())
		if err != nil {
			panic(err)
		}

		fmt.Printf("Looking at round %d\n", round)

		info, err := watcher.AlgodClient.Block(round).Do(context.Background())
		if err != nil {
			panic(err)
		}

		reports := blockAnalyzer.AnalyzeBlock(info)
		discordEvents := []events.Event{}
		telegramEvents := []events.Event{}

		for _, report := range reports {
			event := watcher.Event(report)

			if watcher.ShouldFilterEvent(event) {
				continue
			}

			if watcher.Config.HasDiscordAction(event.EventAction()) {
				discordEvents = append(discordEvents, event)
			}

			if watcher.Config.HasTelegramAction(event.EventAction()) {
				telegramEvents = append(telegramEvents, event)
			}

		}

		if len(discordEvents) > 0 {
			watcher.SendDiscordMessages(discordEvents)
		}

		if len(telegramEvents) > 0 {
			watcher.SendTelegramMessages(telegramEvents)
		}

		round++
	}
}
