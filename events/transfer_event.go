package events

import (
	"fmt"
	"strings"

	"algorillas.com/monko/utils"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TransferEvent struct {
	BaseEvent

	Receiver string
}

func (e TransferEvent) EventAmounts() EventAmounts {
	return EventAmounts{
		AssetAmount: e.AbsAmount,
		AlgoAmount:  e.AlgoAmount,
		USDAmount:   e.USDAmount,
	}
}

func (e TransferEvent) EventAction() Action {
	return e.Action
}

func (e TransferEvent) DiscordEmbed() *discordgo.MessageEmbed {
	abbreviatedSender := utils.AbbreviatedAddress(e.Sender)
	abbreviatedReceiver := utils.AbbreviatedAddress(e.Receiver)
	senderURL := utils.AlloAccountURL(e.Sender)
	receiverURL := utils.AlloAccountURL(e.Receiver)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("[%s](%s) sent %s %s to [%s](%s)",
		abbreviatedSender, senderURL, utils.FormatNumber(e.AbsAmount, 2), e.AssetName, abbreviatedReceiver, receiverURL))

	output.WriteString(fmt.Sprintf("\n\nAlgo Value: %sȺ", utils.FormatNumber(e.AlgoAmount, 2)))
	output.WriteString(fmt.Sprintf("\nUSD Value: $%s", utils.FormatNumber(e.USDAmount, 2)))

	fmt.Printf("%s\n", output.String())

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Transfer", e.AssetName),
		Description: output.String(),

		Type: "rich",
	}

	if e.MediaURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL:    e.MediaURL,
			Width:  e.MediaSize,
			Height: e.MediaSize,
		}
	}

	return embed
}

func (e TransferEvent) TelegramMessage(chatID int64) tgbotapi.VideoConfig {
	return tgbotapi.VideoConfig{}
}
