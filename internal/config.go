package internal

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Discord struct {
		Token    string   `yaml:"token"`
		Channels []string `yaml:"channels"`
	} `yaml:"discord"`

	Algod struct {
		Address string `yaml:"address"`
		Token   string `yaml:"token"`
	} `yaml:"algod"`

	Asset struct {
		ID          uint64  `yaml:"id"`
		Name        string  `yaml:"name"`
		Decimals    uint64  `yaml:"decimals"`
		FilterLimit float64 `yaml:"filter-limit"`
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
