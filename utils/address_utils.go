package utils

func AbbreviatedAddress(address string) string {
	return address[:4] + "..." + address[len(address)-4:]
}
