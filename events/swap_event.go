package events

import (
	"fmt"
	"strings"

	"algorillas.com/monko/utils"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBuyInfo struct {
	NewHolder   bool
	HolderCount int
	Price       float64
	PriceUSD    float64
	ChartURL    string
	WebsiteURL  string
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

	videoConfig := tgbotapi.NewVideo(chatID, tgbotapi.FilePath(e.FilePath()))

	videoConfig.Caption = e.Caption()
	videoConfig.ParseMode = "Markdown"

	return videoConfig

}

func (e SwapEvent) FilePath() string {
	switch {
	case e.NewHolder:
		return "telegram-files/SuperSmp4.mp4"
	case e.USDAmount >= 1000:
		return "telegram-files/bigbuymp4.mp4"
	default:
		return "telegram-files/ExistingHoldersmp4.mp4"
	}
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
	return fmt.Sprintf(`%s
[Buyer](%s)/[Tx](%s)
Amount: %s %s
💸 Value: %s ALGO ($%s USD)
🤲🏽 Holders: %d
💎 Market Cap: %s ALGO
💎 Market Cap: $%s 
📊 [Chart](%s)   💻 [Website](%s)
	`,
		e.EmojiText(),
		utils.AlloAccountURL(e.Sender),
		utils.AlloGroupURL(e.Group),
		utils.FormatNumber(e.AbsAmount, 2), e.AssetName,
		utils.FormatNumber(e.AlgoAmount, 2), utils.FormatNumber(e.USDAmount, 2),
		e.HolderCount,
		utils.FormatNumber(e.TelegramBuyInfo.Price*1e12, 0),
		utils.FormatNumber(e.TelegramBuyInfo.PriceUSD*1e12, 0),
		e.TelegramBuyInfo.ChartURL, e.TelegramBuyInfo.WebsiteURL,
	)
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
