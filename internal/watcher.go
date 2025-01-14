package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/bwmarrin/discordgo"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Watcher struct {
	AssetID   uint64
	AssetName string
	Limit     float64

	AlgodClient *algod.Client
	Bot         *discordgo.Session

	AssetInfoMap map[uint64]AssetInfo
	Channels     []string
}

type Secrets struct {
	DiscordBotToken string
	AlgodAddress    string
	AlgodToken      string
	AssetID         uint64
	AssetName       string
	Limit           float64
	Channels        []string
}

type AssetInfo struct {
	AssetID   uint64
	AssetName string
	Decimals  uint64
}

var (
	p = message.NewPrinter(language.English)
)

func NewWatcher() *Watcher {

	secretData, err := os.ReadFile("secrets.json")
	if err != nil {
		panic(err)
	}

	var secrets Secrets
	if err := json.Unmarshal(secretData, &secrets); err != nil {
		panic(err)
	}

	watcher := &Watcher{
		AssetID:   secrets.AssetID,
		AssetName: secrets.AssetName,
		Limit:     secrets.Limit,
		Channels:  secrets.Channels,
		AssetInfoMap: map[uint64]AssetInfo{
			0: {
				AssetID:   0,
				AssetName: "Algo",
				Decimals:  6,
			},
		},
	}

	if err = json.Unmarshal(secretData, &secrets); err != nil {
		panic(err)
	}

	watcher.Bot, err = discordgo.New("Bot " + secrets.DiscordBotToken)
	if err != nil {
		panic(err)
	}

	watcher.AlgodClient, err = algod.MakeClient(secrets.AlgodAddress, secrets.AlgodToken)
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

	if amt := senderDetails[w.AssetID]; amt != 0 {
		output := strings.Builder{}
		total := math.Abs(float64(amt) / math.Pow10(6))
		abbreviatedSender := fmt.Sprintf("%s...%s", report.Sender[:4], report.Sender[len(report.Sender)-4:])
		senderURL := fmt.Sprintf("https://allo.info/account/%s", report.Sender)
		switch len(senderDetails) {

		case 1:
			receiver := ""
			for wallet := range report.Received {
				receiver = wallet
				break
			}

			abbreviatedReceiver := fmt.Sprintf("%s...%s", receiver[:4], receiver[len(receiver)-4:])
			receiverURL := fmt.Sprintf("https://allo.info/account/%s", receiver)

			output.WriteString(fmt.Sprintf("%s sent %s %s to %s", abbreviatedSender, w.AssetName, formatNumber(total), receiver))
			fmt.Printf("%s\n", output.String())

			if total > w.Limit {
				return &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("%s Transfer", w.AssetName),
					Description: fmt.Sprintf("[%s](%s) sent %s %s to [%s](%s)", abbreviatedSender, senderURL, formatNumber(total), w.AssetName, abbreviatedReceiver, receiverURL),
					Type:        "rich",
				}
			}

		case 2:
			assetID := uint64(0)
			assetAmount := int64(0)
			for asset, aamt := range senderDetails {
				if asset != w.AssetID {
					assetID = asset
					assetAmount = aamt
					break
				}
			}

			info := w.GetAssetInfo(assetID)

			switch strings.HasPrefix(info.AssetName, "TinymanPool") {
			case true:
				op := "REMOVED"
				if amt < 0 {
					op = "ADDED"
				}
				output.WriteString(fmt.Sprintf("[%s](%s) %s %s of liquidity in %s", abbreviatedSender, senderURL, op, formatNumber(total), info.AssetName))
			case false:
				op := "BOUGHT"
				if amt < 0 {
					op = "SOLD"
				}
				output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s for %s %s", abbreviatedSender, senderURL,
					op, formatNumber(total), w.AssetName, formatNumber(math.Abs(float64(assetAmount)/math.Pow10(int(info.Decimals)))), info.AssetName))
			}

			fmt.Printf("%s\n", output.String())
			if total > w.Limit {
				groupURL := fmt.Sprintf("https://allo.info/tx/group/%s", url.QueryEscape(report.Group))

				return &discordgo.MessageEmbed{
					Title:       fmt.Sprintf("%s Swap", w.AssetName),
					Description: output.String(),
					URL:         groupURL,
					Type:        "link",
				}
			}

		case 3:
			poolID := uint64(0)
			otherAssetID := uint64(0)
			otherAssetAmount := int64(0)

			for asset, aamt := range senderDetails {
				if asset != w.AssetID {
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
				if amt < 0 {
					op = "ADDED"
				}
				otherInfo := w.GetAssetInfo(otherAssetID)
				poolInfo := w.GetAssetInfo(poolID)

				output.WriteString(fmt.Sprintf("[%s](%s) %s %s %s and %s %s of liquidity in %s", abbreviatedSender, senderURL, op, formatNumber(total), w.AssetName,
					formatNumber(math.Abs(float64(otherAssetAmount)/math.Pow10(int(otherInfo.Decimals)))), otherInfo.AssetName, poolInfo.AssetName))
				fmt.Printf("%s\n", output.String())

				if total > w.Limit {

					groupURL := fmt.Sprintf("https://allo.info/tx/group/%s", url.QueryEscape(report.Group))

					return &discordgo.MessageEmbed{
						Title:       fmt.Sprintf("%s Swap", w.AssetName),
						Description: output.String(),
						URL:         groupURL,
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
	for _, channel := range w.Channels {
		for _, message := range messages {
			w.Bot.ChannelMessageSendEmbed(channel, message)
		}
	}
}

func formatNumber(num float64) string {
	return p.Sprintf("%.2f", num)
}
