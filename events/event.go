package events

import (
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramResult struct {
	VideoConfig tgbotapi.VideoConfig
}

type EventAmounts struct {
	AssetAmount float64
	AlgoAmount  float64
	USDAmount   float64
}

type Event interface {
	EventAction() Action
	EventAmounts() EventAmounts
	DiscordEmbed() *discordgo.MessageEmbed
	TelegramMessage(chatID int64) tgbotapi.VideoConfig
}

type BaseEvent struct {
	Sender string
	Action Action

	AssetName string
	AssetID   uint64
	Amount    float64
	AbsAmount float64

	AlgoAmount float64
	USDAmount  float64
	USDAssetID uint64

	MediaURL  string
	MediaSize int
}
