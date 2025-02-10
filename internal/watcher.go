package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"

	"algorillas.com/monko/config"
	"algorillas.com/monko/events"
	"algorillas.com/monko/utils"
	"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"github.com/bwmarrin/discordgo"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AssetInfo struct {
	AssetID   uint64
	AssetName string
	Decimals  uint64
}

type Watcher struct {
	Config config.Config

	AlgodClient *algod.Client
	DiscordBot  *discordgo.Session
	TelegramBot *tgbotapi.BotAPI

	AssetPrice float64
	AlgoPrice  float64

	AssetInfoMap map[uint64]AssetInfo

	HolderCount int
}

func NewWatcher() *Watcher {

	config, err := config.GetConfigFromFile("config.yaml")
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
				Decimals:  config.Asset.Decimals,
			},
		},
	}

	watcher.DiscordBot, err = discordgo.New("Bot " + config.Discord.Token)
	if err != nil {
		panic(err)
	}

	watcher.TelegramBot, err = tgbotapi.NewBotAPI(config.Telegram.Token)
	watcher.TelegramBot.Debug = true
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
		fmt.Printf("Could not find asset %d", assetID)
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

func (w *Watcher) Event(report *AssetReport) events.Event {
	report.CalculateDeltas()
	senderDetails := report.Deltas[report.Sender]

	var accountBalace float64

	balanceInfo, err := w.AlgodClient.AccountAssetInformation(report.Sender, w.Config.Asset.ID).Do(context.Background())
	switch err != nil {
	case true:
		accountBalace = 0
	case false:
		accountBalace = float64(balanceInfo.AssetHolding.Amount) / math.Pow10(int(w.Config.Asset.Decimals))
	}

	if amt := senderDetails[w.Config.Asset.ID]; amt != 0 {

		total := float64(amt) / math.Pow10(int(w.Config.Asset.Decimals))

		baseEvent := events.BaseEvent{
			Sender: report.Sender,

			AssetID:   w.Config.Asset.ID,
			AssetName: w.Config.Asset.Name,
			Amount:    total,
			AbsAmount: math.Abs(total),

			AlgoAmount: math.Abs(total) * w.AssetPrice,
			USDAmount:  math.Abs(total) * w.AssetPrice * w.AlgoPrice,
			USDAssetID: w.Config.Price.Usd.ID,

			MediaSize: w.Config.Image.Size,
		}

		var filterAmount float64
		switch w.Config.Asset.FilterAsset {
		case "ALGO":
			filterAmount = baseEvent.AlgoAmount
		case "USD":
			filterAmount = baseEvent.USDAmount
		default:
			filterAmount = baseEvent.AbsAmount
		}

		switch len(senderDetails) {

		case 1:
			receiver := ""
			for wallet := range report.Received {
				receiver = wallet
				break
			}

			baseEvent.Action = events.TransferAction
			baseEvent.MediaURL = w.Config.Image.TransferURL

			return events.TransferEvent{
				BaseEvent: baseEvent,
				Receiver:  receiver,
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
			assetTotal := float64(assetAmount) / math.Pow10(int(info.Decimals))

			if utils.Signum(total) == utils.Signum(assetTotal) {
				switch total > 0 {
				case true:
					baseEvent.Action = events.WithdrawAction
				case false:
					baseEvent.Action = events.DepositAction
				}

				baseEvent.MediaURL = w.Config.Image.TransferURL

				return events.ContractEvent{
					BaseEvent: baseEvent,

					Group: report.Group,
				}
			}

			switch strings.HasPrefix(info.AssetName, "TinymanPool") {
			case true:

				switch amt < 0 {
				case true:
					baseEvent.Action = events.AddAction
				case false:
					baseEvent.Action = events.RemoveAction
				}

				baseEvent.MediaURL = w.Config.ImageURL(baseEvent.Action, filterAmount)

				return events.LiquidityEvent{
					BaseEvent: baseEvent,

					Group: report.Group,

					PoolAssetID:   assetID,
					PoolAssetName: info.AssetName,
					PoolAmount:    assetTotal,
					PoolAbsAmount: math.Abs(assetTotal),
				}

			case false:

				switch amt < 0 {
				case true:
					baseEvent.Action = events.SellAction
				case false:
					baseEvent.Action = events.BuyAction
				}

				baseEvent.MediaURL = w.Config.ImageURL(baseEvent.Action, filterAmount)

				swap := &events.SwapEvent{
					BaseEvent: baseEvent,

					Group: report.Group,

					ToAssetID:   assetID,
					ToAssetName: info.AssetName,
					ToAmount:    assetTotal,
					ToAbsAmount: math.Abs(assetTotal),
				}

				if baseEvent.Action == events.BuyAction {
					newHolder := math.Abs(total-accountBalace) < 0.0001

					var videoPath string

					switch newHolder {
					case true:
						videoPath = w.Config.Telegram.Videos.NewHolder
					case false:
						videoPath = w.Config.Telegram.Videos.ExistingHolder
					}

					if baseEvent.USDAmount >= w.Config.Telegram.LargeBuyLimit {
						videoPath = w.Config.Telegram.Videos.LargeBuy
					}

					swap.TelegramBuyInfo = events.TelegramBuyInfo{
						NewHolder:         newHolder,
						HolderCount:       w.HolderCount,
						Price:             w.AssetPrice,
						PriceUSD:          w.AssetPrice * w.AlgoPrice,
						Tokens:            w.Config.Asset.Tokens,
						ChartURL:          w.Config.Asset.ChartURL,
						WebsiteURL:        w.Config.Asset.Website,
						TelegramVideoPath: videoPath,
					}

				}

				return *swap
			}

		case 3:
			poolID := uint64(0)
			poolAmount := int64(0)
			otherAssetID := uint64(0)
			otherAssetAmount := int64(0)

			for asset, aamt := range senderDetails {
				if asset != w.Config.Asset.ID {
					info := w.GetAssetInfo(asset)

					switch strings.HasPrefix(info.AssetName, "TinymanPool") || strings.Contains(info.AssetName, "Pact") {
					case true:
						poolID = asset
						poolAmount = aamt
					default:
						otherAssetID = asset
						otherAssetAmount = aamt
					}
				}
			}

			if poolID != 0 {

				switch amt < 0 {
				case true:
					baseEvent.Action = events.AddAction
				case false:
					baseEvent.Action = events.RemoveAction
				}

				otherInfo := w.GetAssetInfo(otherAssetID)
				poolInfo := w.GetAssetInfo(poolID)

				poolTotal := float64(poolAmount) / math.Pow10(int(poolInfo.Decimals))
				otherTotal := float64(otherAssetAmount) / math.Pow10(int(otherInfo.Decimals))

				baseEvent.MediaURL = w.Config.ImageURL(baseEvent.Action, filterAmount)

				return events.LiquidityEvent{
					BaseEvent: baseEvent,

					Group: report.Group,

					PairAssetID:   otherAssetID,
					PairAssetName: otherInfo.AssetName,
					PairAmount:    otherTotal,
					PairAbsAmount: math.Abs(otherTotal),

					PoolAssetID:   poolID,
					PoolAssetName: poolInfo.AssetName,
					PoolAmount:    poolTotal,
					PoolAbsAmount: math.Abs(poolTotal),
				}

			} else {
				// Debugging purposes
				deltaOutput, _ := json.MarshalIndent(report.Deltas, "", "  ")
				fmt.Printf("%s made a complex transaction\n%s\n", report.Sender, string(deltaOutput))
			}

		default:
			// Debugging purposes
			deltaOutput, _ := json.MarshalIndent(report.Deltas, "", "  ")
			fmt.Printf("%s made a complex transaction\n%s\n", report.Sender, string(deltaOutput))
		}

	}
	return nil
}

func (w *Watcher) SendDiscordMessages(events []events.Event) {
	for _, event := range events {
		for _, channel := range w.Config.Discord.Channels {
			w.DiscordBot.ChannelMessageSendEmbed(channel, event.DiscordEmbed())
		}
	}

}

func (w *Watcher) SendTelegramMessages(events []events.Event) {

	for _, event := range events {
		for _, chatID := range w.Config.Telegram.ChatIDs {
			w.TelegramBot.Send(event.TelegramMessage(chatID))
		}
	}

}

func (w *Watcher) ShouldFilterEvent(event events.Event) bool {
	if event == nil {
		return true
	}

	amounts := event.EventAmounts()

	switch w.Config.Asset.FilterAsset {
	case "ALGO":
		return amounts.AlgoAmount < w.Config.Asset.FilterLimit
	case "ASSET":
		return amounts.AssetAmount < w.Config.Asset.FilterLimit
	case "USD":
		return amounts.USDAmount < w.Config.Asset.FilterLimit
	}

	return true
}

func (w *Watcher) CalcPrices(startRound uint64, currentRound uint64) {
	if !w.Config.Price.Track {
		return
	}

	if (currentRound-startRound)%w.Config.Price.BlockInterval == 0 {
		w.CalcAssetPrice()
	}

	if (currentRound-startRound)%w.Config.Price.Usd.BlockInterval == 0 {
		w.CalcAlgoPrice()
	}

}

func (w *Watcher) CalcAssetPrice() {
	accountInfo, err := w.AlgodClient.AccountInformation(w.Config.Price.PrimaryAlgoLpAddress).Do(context.Background())
	if err != nil {
		panic(err)
	}

	algoAmount := accountInfo.AmountWithoutPendingRewards

	assetAmount, err := w.AlgodClient.AccountAssetInformation(w.Config.Price.PrimaryAlgoLpAddress,
		w.Config.Asset.ID).Do(context.Background())
	if err != nil {
		panic(err)
	}

	assetInfo := w.GetAssetInfo(w.Config.Asset.ID)

	assetAmountFloat := float64(assetAmount.AssetHolding.Amount) / math.Pow10(int(assetInfo.Decimals))
	algoAmountFloat := float64(algoAmount) / math.Pow10(6)

	w.AssetPrice = algoAmountFloat / assetAmountFloat
	fmt.Printf("Asset Price: %.08fȺ\n", w.AssetPrice)
}

func (w *Watcher) CalcAlgoPrice() {
	accountInfo, err := w.AlgodClient.AccountInformation(w.Config.Price.Usd.PrimaryAlgoLPAddress).Do(context.Background())
	if err != nil {
		panic(err)
	}

	algoAmount := accountInfo.AmountWithoutPendingRewards

	assetAmount, err := w.AlgodClient.AccountAssetInformation(w.Config.Price.Usd.PrimaryAlgoLPAddress,
		w.Config.Price.Usd.ID).Do(context.Background())
	if err != nil {
		panic(err)
	}

	assetAmountFloat := float64(assetAmount.AssetHolding.Amount) / math.Pow10(6)
	algoAmountFloat := float64(algoAmount) / math.Pow10(6)

	w.AlgoPrice = assetAmountFloat / algoAmountFloat
	fmt.Printf("Algo Price: $%.03f\n", w.AlgoPrice)
}

func (w *Watcher) UpdateHolderCount(startRound uint64, currentRound uint64) {
	// Intentionally hard-coded to only look at the holder count every 1333 rounds
	// This is approximately once per hour.  Do not want to spam the API
	if (currentRound-startRound)%1333 == 0 {
		req, err := w.getHolderCountRequest()
		if err != nil {
			panic(err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		response := struct {
			Holders struct {
				TotalCount int `json:"totalCount"`
			} `json:"holders"`
		}{}

		err = json.Unmarshal(body, &response)
		if err != nil {
			panic(err)
		}

		w.HolderCount = response.Holders.TotalCount
	}
}

func (w *Watcher) getHolderCountRequest() (*http.Request, error) {
	url := "https://allo.info/api/v1/graphql/getAssetHoldersCount"

	var jsonStr = []byte(fmt.Sprintf(`{"id": %d}`, w.Config.Asset.ID))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referrer", fmt.Sprintf("https://allo.info/asset/%d/holders", w.Config.Asset.ID))
	req.Header.Set("Referrer-Policy", "unsafe-url")
	req.Header.Set("Origin", "https://allo.info")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")

	return req, nil
}
