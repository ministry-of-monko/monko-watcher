package events

import (
	"fmt"
	"strings"

	"algorillas.com/monko/utils"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type LiquidityEvent struct {
	BaseEvent

	Group string

	PairAssetID   uint64
	PairAssetName string
	PairAmount    float64
	PairAbsAmount float64

	PoolAssetID   uint64
	PoolAssetName string
	PoolAmount    float64
	PoolAbsAmount float64
}

func (e LiquidityEvent) EventAmounts() EventAmounts {
	return EventAmounts{
		AssetAmount: e.AbsAmount,
		AlgoAmount:  e.AlgoAmount,
		USDAmount:   e.USDAmount,
	}
}

func (e LiquidityEvent) EventAction() Action {
	return e.Action
}

func (e LiquidityEvent) DiscordEmbed() *discordgo.MessageEmbed {
	abbreviatedSender := utils.AbbreviatedAddress(e.Sender)
	senderURL := utils.AlloAccountURL(e.Sender)

	op := ""

	switch e.Action {
	case AddAction:
		op = "ADDED"
	case RemoveAction:
		op = "REMOVED"
	}

	var output strings.Builder
	other := ""
	if e.PairAbsAmount > 0 {
		other = fmt.Sprintf(" and %s %s", utils.FormatNumber(e.PairAbsAmount, 2), e.PairAssetName)
	}

	output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s%s to %s\n",
		abbreviatedSender, senderURL, op, utils.FormatNumber(e.AbsAmount, 2), e.AssetName, other, e.PoolAssetName))

	if e.PairAssetID != 0 {
		output.WriteString(fmt.Sprintf("\nAlgo Value: %sȺ", utils.FormatNumber(e.AlgoAmount, 2)))
	}

	if e.PairAssetID != e.USDAssetID {
		output.WriteString(fmt.Sprintf("\nUSD Value: $%s", utils.FormatNumber(e.USDAmount, 2)))
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Liquidity Event", e.AssetName),
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

func (e LiquidityEvent) TelegramMessage(chatID int64) tgbotapi.VideoConfig {
	return tgbotapi.VideoConfig{}
}
