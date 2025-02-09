package config

import (
	"os"

	"algorillas.com/monko/events"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Discord struct {
		Token    string   `yaml:"token"`
		Actions  []string `yaml:"actions"`
		Channels []string `yaml:"channels"`
	} `yaml:"discord"`

	Telegram struct {
		Token   string   `yaml:"token"`
		Actions []string `yaml:"actions"`
		ChatIDs []int64  `yaml:"chat-ids"`
	} `yaml:"telegram"`

	Algod struct {
		Address string `yaml:"address"`
		Token   string `yaml:"token"`
	} `yaml:"algod"`

	Asset struct {
		ID             uint64  `yaml:"id"`
		Name           string  `yaml:"name"`
		Decimals       uint64  `yaml:"decimals"`
		FilterLimit    float64 `yaml:"filter-limit"`
		FilterAsset    string  `yaml:"filter-asset"`
		HolderInterval uint64  `yaml:"holder-interval"`
		ChartURL       string  `yaml:"chart-url"`
		Website        string  `yaml:"website"`
	} `yaml:"asset"`

	Price struct {
		Track                bool   `yaml:"track"`
		BlockInterval        uint64 `yaml:"block-interval"`
		PrimaryAlgoLpAddress string `yaml:"primary-algo-lp-address"`
		Usd                  struct {
			ID                   uint64 `yaml:"id"`
			PrimaryAlgoLPAddress string `yaml:"primary-algo-lp-address"`
			BlockInterval        uint64 `yaml:"block-interval"`
		} `yaml:"usd"`
	} `yaml:"price"`

	Image struct {
		Size               int    `yaml:"size"`
		TransferURL        string `yaml:"transfer-url"`
		LiquidityAddURL    string `yaml:"liquidity-add-url"`
		LiquidityRemoveURL string `yaml:"liquidity-remove-url"`
		Buy                []struct {
			Limit float64 `yaml:"limit"`
			URL   string  `yaml:"url"`
		} `yaml:"buy"`
		Sell []struct {
			Limit float64 `yaml:"limit"`
			URL   string  `yaml:"url"`
		} `yaml:"sell"`
	} `yaml:"image"`

	TelegramVideos struct {
		NewHolderURL      string `yaml:"new-holder"`
		ExistingHolderURL string `yaml:"existing-holder"`
		LargeBuyURL       string `yaml:"large-buy"`
	} `yaml:"telegram-videos"`
}

func GetConfigFromFile(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}

	if err = yaml.Unmarshal(data, &config); err != nil {
		return config, err
	}

	return config, nil
}

func (c Config) ImageURL(action events.Action, amount float64) string {
	var url string

	switch action {

	case events.AddAction:
		url = c.Image.LiquidityAddURL

	case events.RemoveAction:
		url = c.Image.LiquidityRemoveURL

	case events.BuyAction:
		for _, possibility := range c.Image.Buy {
			if amount >= possibility.Limit {
				url = possibility.URL
			}
		}
	case events.SellAction:
		for _, possibility := range c.Image.Sell {
			if amount >= possibility.Limit {
				url = possibility.URL
			}
		}
	}

	return url
}

func (c Config) HasDiscordAction(action events.Action) bool {
	for _, a := range c.Discord.Actions {
		if a == string(action) {
			return true
		}
	}

	return false
}

func (c Config) HasTelegramAction(action events.Action) bool {
	for _, a := range c.Telegram.Actions {
		if a == string(action) {
			return true
		}
	}

	return false
}
