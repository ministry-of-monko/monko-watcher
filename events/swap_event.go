package events

import (
	"fmt"
	"strings"

	"algorillas.com/monko/utils"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBuyInfo struct {
	NewHolder         bool
	HolderCount       int
	Price             float64
	PriceUSD          float64
	Tokens            float64
	ChartURL          string
	WebsiteURL        string
	TelegramVideoPath string
}

type SwapEvent struct {
	BaseEvent
	TelegramBuyInfo

	Group string

	ToAssetID   uint64
	ToAssetName string
	ToAmount    float64
	ToAbsAmount float64
}

func (e SwapEvent) EventAmounts() EventAmounts {
	return EventAmounts{
		AssetAmount: e.AbsAmount,
		AlgoAmount:  e.AlgoAmount,
		USDAmount:   e.USDAmount,
	}
}

func (e SwapEvent) EventAction() Action {
	return e.Action
}

func (e SwapEvent) DiscordEmbed() *discordgo.MessageEmbed {

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Swap", e.AssetName),
		Description: e.Description(),
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

func (e SwapEvent) TelegramMessage(chatID int64) tgbotapi.VideoConfig {

	videoConfig := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(fmt.Sprintf("telegram-files/%s", e.TelegramVideoPath)))

	videoConfig.Caption = e.Caption()
	videoConfig.ParseMode = "Markdown"

	return videoConfig

}

func (e SwapEvent) Description() string {
	var output strings.Builder

	abbreviatedSender := utils.AbbreviatedAddress(e.Sender)
	senderURL := utils.AlloAccountURL(e.Sender)

	op := ""

	switch e.Action {
	case BuyAction:
		op = "BOUGHT"
	case SellAction:
		op = "SOLD"
	}
	output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s for %s %s\n",
		abbreviatedSender, senderURL, op, utils.FormatNumber(e.AbsAmount, 2), e.AssetName, utils.FormatNumber(e.ToAbsAmount, 2), e.ToAssetName))

	if e.ToAssetID != 0 {
		output.WriteString(fmt.Sprintf("\nAlgo Value: %sȺ", utils.FormatNumber(e.AlgoAmount, 2)))
	}

	if e.ToAssetID != e.USDAssetID {
		output.WriteString(fmt.Sprintf("\nUSD Value: $%s", utils.FormatNumber(e.USDAmount, 2)))
	}

	return output.String()
}

func (e SwapEvent) Caption() string {

	algoAmount := e.AlgoAmount
	if e.ToAssetID == 0 {
		algoAmount = e.ToAbsAmount
	}

	var output strings.Builder

	output.WriteString(fmt.Sprintf("%s\n", e.EmojiText()))
	output.WriteString(fmt.Sprintf("[Buyer](%s) / [Tx](%s)\n", utils.AlloAccountURL(e.Sender), utils.AlloGroupURL(e.Group)))
	output.WriteString(fmt.Sprintf("Amount: %s %s\n", utils.FormatNumber(e.AbsAmount, 2), e.AssetName))
	output.WriteString(fmt.Sprintf("💸 Value: %s ALGO\n", utils.FormatNumber(algoAmount, 2)))
	output.WriteString(fmt.Sprintf("💸 Value: $%s USD\n", utils.FormatNumber(e.USDAmount, 2)))
	output.WriteString(fmt.Sprintf("🤲🏽 Holders: %d\n", e.HolderCount))
	output.WriteString(fmt.Sprintf("💎 Market Cap: $%s\n", utils.FormatNumber(e.TelegramBuyInfo.PriceUSD*e.Tokens, 0)))
	output.WriteString(fmt.Sprintf("📊 [Chart](%s)   💻 [Website](%s)\n", e.TelegramBuyInfo.ChartURL, e.TelegramBuyInfo.WebsiteURL))

	return output.String()
}

func (e SwapEvent) EmojiText() string {
	switch {
	case e.NewHolder:
		return "🚀🔥🚀🔥🚀🔥🚀🔥🚀🔥🚀\n🔥🚀🔥🚀🔥🚀🔥🚀🔥🚀🔥\n⬆️ New holder!"
	case e.USDAmount >= 1000:
		return "🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽\n🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧\n🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽\n🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧\n⬆️ Existing holder!"
	default:
		return "🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽\n🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧🫵🏽🦧\n⬆️ Existing holder!"
	}
}
