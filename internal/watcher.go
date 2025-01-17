package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Action string

type AssetInfo struct {
	AssetID   uint64
	AssetName string
	Decimals  uint64
}

type Watcher struct {
	Config Config

	AlgodClient *algod.Client
	Bot         *discordgo.Session

	AssetPrice float64
	AlgoPrice  float64

	AssetInfoMap map[uint64]AssetInfo
}

const (
	USDCLPAddress = "2PIFZW53RHCSFSYMCFUBW4XOCXOMB7XOYQSQ6KGT3KVGJTL4HM6COZRNMM"
	USDCAssetID   = 31566704

	TransferAction = Action("transfer")
	BuyAction      = Action("bought")
	SellAction     = Action("sold")
	AddAction      = Action("added")
	RemoveAction   = Action("removed")
)

var (
	p = message.NewPrinter(language.English)
)

func NewWatcher() *Watcher {

	config, err := GetConfigFromFile("config.yaml")
	if err != nil {
		panic(err)
	}

	assetID := uint64(config.Asset.ID)

	watcher := &Watcher{
		Config: config,
		AssetInfoMap: map[uint64]AssetInfo{
			0: {
				AssetID:   0,
				AssetName: "Algo",
				Decimals:  6,
			},
			assetID: {
				AssetID:   assetID,
				AssetName: config.Asset.Name,
				Decimals:  uint64(config.Asset.Decimals),
			},
		},
	}

	watcher.Bot, err = discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		panic(err)
	}

	watcher.AlgodClient, err = algod.MakeClient(config.Algod.Address, config.Algod.Token)
	if err != nil {
		panic(err)
	}

	return watcher
}

func (w *Watcher) GetAssetInfo(assetID uint64) AssetInfo {
	assetInfo, found := w.AssetInfoMap[assetID]
	if found {
		return assetInfo
	}

	info, err := w.AlgodClient.GetAssetByID(assetID).Do(context.Background())
	if err != nil {
		panic(err)
	}

	assetInfo = AssetInfo{
		AssetID:   assetID,
		AssetName: info.Params.Name,
		Decimals:  info.Params.Decimals,
	}

	w.AssetInfoMap[assetID] = assetInfo

	return assetInfo
}

func (w *Watcher) GetDiscordEmbedFromReport(report *AssetReport) *discordgo.MessageEmbed {
	report.CalculateDeltas()
	senderDetails := report.Deltas[report.Sender]

	if amt := senderDetails[w.Config.Asset.ID]; amt != 0 {
		output := strings.Builder{}
		total := math.Abs(float64(amt) / math.Pow10(6))
		abbreviatedSender := fmt.Sprintf("%s...%s", report.Sender[:4], report.Sender[len(report.Sender)-4:])
		senderURL := fmt.Sprintf("https://allo.info/account/%s", report.Sender)
		var action Action

		switch len(senderDetails) {

		case 1:
			receiver := ""
			for wallet := range report.Received {
				receiver = wallet
				break
			}

			abbreviatedReceiver := fmt.Sprintf("%s...%s", receiver[:4], receiver[len(receiver)-4:])
			receiverURL := fmt.Sprintf("https://allo.info/account/%s", receiver)

			output.WriteString(fmt.Sprintf("[%s](%s) sent %s %s to [%s](%s)", abbreviatedSender, senderURL, formatNumber(total), w.Config.Asset.Name, abbreviatedReceiver, receiverURL))
			output.WriteString(fmt.Sprintf("\n\nAlgo Value: %s", formatNumber(total*w.AssetPrice)))
			output.WriteString(fmt.Sprintf("\nUSD Value: $%s", formatNumber(total*w.AssetPrice*w.AlgoPrice)))

			fmt.Printf("%s\n", output.String())

			if total > w.Config.Asset.FilterLimit {

				thumb := w.GetEmbedThumbnail(TransferAction, total)
				return &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("%s Transfer", w.Config.Asset.Name),
					Description: output.String(),
					Thumbnail:   thumb,
					Type:        "rich",
				}
			}

		case 2:
			assetID := uint64(0)
			assetAmount := int64(0)
			for asset, aamt := range senderDetails {
				if asset != w.Config.Asset.ID {
					assetID = asset
					assetAmount = aamt
					break
				}
			}

			info := w.GetAssetInfo(assetID)

			switch strings.HasPrefix(info.AssetName, "TinymanPool") {
			case true:
				op := "REMOVED"
				action = RemoveAction
				if amt < 0 {
					op = "ADDED"
					action = AddAction
				}
				output.WriteString(fmt.Sprintf("[%s](%s) %s %s of liquidity in %s", abbreviatedSender, senderURL, op, formatNumber(total), info.AssetName))
			case false:
				op := "BOUGHT"
				action = BuyAction
				if amt < 0 {
					op = "SOLD"
					action = SellAction
				}
				output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s for %s %s\n", abbreviatedSender, senderURL,
					op, formatNumber(total), w.Config.Asset.Name, formatNumber(math.Abs(float64(assetAmount)/math.Pow10(int(info.Decimals)))), info.AssetName))

				if assetID != 0 {
					output.WriteString(fmt.Sprintf("\nAlgo Value: %s", formatNumber(total*w.AssetPrice)))
				}

				output.WriteString(fmt.Sprintf("\nUSD Value: $%s", formatNumber(total*w.AssetPrice*w.AlgoPrice)))
			}

			fmt.Printf("%s\n", output.String())
			if total > w.Config.Asset.FilterLimit {
				thumb := w.GetEmbedThumbnail(action, total)
				groupURL := fmt.Sprintf("https://allo.info/tx/group/%s", url.QueryEscape(report.Group))

				return &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("%s Swap", w.Config.Asset.Name),
					Description: output.String(),
					URL:         groupURL,
					Thumbnail:   thumb,
					Type:        "link",
				}
			}

		case 3:
			poolID := uint64(0)
			otherAssetID := uint64(0)
			otherAssetAmount := int64(0)

			for asset, aamt := range senderDetails {
				if asset != w.Config.Asset.ID {
					info := w.GetAssetInfo(asset)

					switch strings.HasPrefix(info.AssetName, "TinymanPool") || strings.Contains(info.AssetName, "Pact") {
					case true:
						poolID = asset
					default:
						otherAssetID = asset
						otherAssetAmount = aamt
					}
				}
			}

			if poolID != 0 {
				op := "REMOVED"
				action = RemoveAction
				if amt < 0 {
					op = "ADDED"
					action = AddAction
				}
				otherInfo := w.GetAssetInfo(otherAssetID)
				poolInfo := w.GetAssetInfo(poolID)

				output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s and %s %s of liquidity in %s", abbreviatedSender, senderURL, op, formatNumber(total), w.Config.Asset.Name,
					formatNumber(math.Abs(float64(otherAssetAmount)/math.Pow10(int(otherInfo.Decimals)))), otherInfo.AssetName, poolInfo.AssetName))
				fmt.Printf("%s\n", output.String())

				if total > w.Config.Asset.FilterLimit {

					groupURL := fmt.Sprintf("https://allo.info/tx/group/%s", url.QueryEscape(report.Group))
					thumb := w.GetEmbedThumbnail(action, total)

					return &discordgo.MessageEmbed{
						Title:       fmt.Sprintf("%s Swap", w.Config.Asset.Name),
						Description: output.String(),
						URL:         groupURL,
						Thumbnail:   thumb,
						Type:        "link",
					}
				}
			} else {
				// Debugging purposes
				deltaOutput, _ := json.MarshalIndent(report.Deltas, "", "  ")
				output.WriteString(fmt.Sprintf("%s made a complex transaction\n%s\n", report.Sender, string(deltaOutput)))
			}

		default:
			// Debugging purposes
			deltaOutput, _ := json.MarshalIndent(report.Deltas, "", "  ")
			output.WriteString(fmt.Sprintf("%s made a complex transaction\n%s\n", report.Sender, string(deltaOutput)))
		}

		fmt.Printf("%s\n", output.String())

	}
	return nil
}

func (w *Watcher) SendMessages(messages []*discordgo.MessageEmbed) {
	for _, channel := range w.Config.Discord.Channels {
		for _, message := range messages {
			w.Bot.ChannelMessageSendEmbed(channel, message)
		}
	}
}

func (w *Watcher) CalcAssetPrice() {
	accountInfo, err := w.AlgodClient.AccountInformation(w.Config.Asset.PrimaryAlgoLPAddress).Do(context.Background())
	if err != nil {
		panic(err)
	}

	algoAmount := accountInfo.AmountWithoutPendingRewards

	assetAmount, err := w.AlgodClient.AccountAssetInformation(w.Config.Asset.PrimaryAlgoLPAddress, w.Config.Asset.ID).Do(context.Background())
	if err != nil {
		panic(err)
	}

	assetInfo := w.GetAssetInfo(w.Config.Asset.ID)

	assetAmountFloat := float64(assetAmount.AssetHolding.Amount) / math.Pow10(int(assetInfo.Decimals))
	algoAmountFloat := float64(algoAmount) / math.Pow10(6)

	w.AssetPrice = algoAmountFloat / assetAmountFloat
}

func (w *Watcher) CalcAlgoPrice() {
	accountInfo, err := w.AlgodClient.AccountInformation(USDCLPAddress).Do(context.Background())
	if err != nil {
		panic(err)
	}

	algoAmount := accountInfo.AmountWithoutPendingRewards

	assetAmount, err := w.AlgodClient.AccountAssetInformation(USDCLPAddress, USDCAssetID).Do(context.Background())
	if err != nil {
		panic(err)
	}

	assetAmountFloat := float64(assetAmount.AssetHolding.Amount) / math.Pow10(6)
	algoAmountFloat := float64(algoAmount) / math.Pow10(6)

	w.AlgoPrice = assetAmountFloat / algoAmountFloat
}

func (w *Watcher) GetEmbedThumbnail(action Action, amount float64) *discordgo.MessageEmbedThumbnail {
	if len(action) == 0 {
		return nil
	}

	var url string

	switch action {

	case TransferAction:
		url = w.Config.Image.TransferURL

	case AddAction:
		url = w.Config.Image.LiquidityAddURL

	case RemoveAction:
		url = w.Config.Image.LiquidityRemoveURL

	case BuyAction:
		for _, possibility := range w.Config.Image.Buy {
			if amount >= possibility.Limit {
				url = possibility.URL
			}
		}
	case SellAction:
		for _, possibility := range w.Config.Image.Sell {
			if amount >= possibility.Limit {
				url = possibility.URL
			}
		}
	}

	if len(url) == 0 {
		return nil
	}

	return &discordgo.MessageEmbedThumbnail{
		URL:    url,
		Width:  w.Config.Image.Size,
		Height: w.Config.Image.Size,
	}
}

func formatNumber(num float64) string {
	return p.Sprintf("%.2f", num)
}
