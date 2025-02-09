package events

import (
	"fmt"
	"strings"

	"algorillas.com/monko/utils"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ContractEvent struct {
	BaseEvent

	Group string
}

func (e ContractEvent) EventAmounts() EventAmounts {
	return EventAmounts{
		AssetAmount: e.AbsAmount,
		AlgoAmount:  e.AlgoAmount,
		USDAmount:   e.USDAmount,
	}
}

func (e ContractEvent) EventAction() Action {
	return e.Action
}

func (e ContractEvent) DiscordEmbed() *discordgo.MessageEmbed {
	abbreviatedSender := utils.AbbreviatedAddress(e.Sender)
	senderURL := utils.AlloAccountURL(e.Sender)

	op := ""
	dir := ""

	switch e.Action {
	case DepositAction:
		op = "DEPOSITED"
		dir = "into"
	case WithdrawAction:
		op = "WITHDREW"
		dir = "from"
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s %s a contract",
		abbreviatedSender, senderURL, op, utils.FormatNumber(e.AbsAmount, 2), e.AssetName, dir))

	output.WriteString(fmt.Sprintf("\n\nAlgo Value: %sȺ", utils.FormatNumber(e.AlgoAmount, 2)))
	output.WriteString(fmt.Sprintf("\nUSD Value: $%s", utils.FormatNumber(e.USDAmount, 2)))

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Contract Transaction", e.AssetName),
		Description: output.String(),
		URL:         utils.AlloGroupURL(e.Group),
		Type:        "link",
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

func (e ContractEvent) TelegramMessage(chatID int64) tgbotapi.VideoConfig {
	return tgbotapi.VideoConfig{}
}
