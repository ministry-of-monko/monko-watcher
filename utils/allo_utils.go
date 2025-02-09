package utils

import (
	"fmt"
	"net/url"
)

func AlloAccountURL(address string) string {
	return fmt.Sprintf("https://allo.info/account/%s", address)
}

func AlloGroupURL(group string) string {
	return fmt.Sprintf("https://allo.info/tx/group/%s", url.QueryEscape(group))
}
