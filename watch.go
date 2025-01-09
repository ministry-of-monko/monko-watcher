package main

import (
	"context"
	"fmt"

	"algorillas.com/monko/internal"
	"github.com/bwmarrin/discordgo"
)

func main() {

	watcher := internal.NewWatcher()

	status, err := watcher.AlgodClient.Status().Do(context.Background())
	if err != nil {
		panic(err)
	}

	round := status.LastRound

	blockAnalyzer := internal.NewBlockAnalyzer(watcher.AssetID)

	fmt.Printf("%v", watcher)

	for {
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