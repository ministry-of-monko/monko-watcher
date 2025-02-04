package main

import (
	"context"
	"fmt"

	"algorillas.com/monko/internal"
	"github.com/bwmarrin/discordgo"
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
		embeds := []*discordgo.MessageEmbed{}

		for _, report := range reports {
			embed := watcher.GetDiscordEmbedFromReport(report)

			if embed != nil {
				embeds = append(embeds, embed)
			}
		}

		watcher.SendMessages(embeds)
		round++
	}
}
