package internal

type Action string

const (
	TransferAction = Action("transfer")
	BuyAction      = Action("buy")
	SellAction     = Action("sell")
	AddAction      = Action("add")
	RemoveAction   = Action("remove")
)
