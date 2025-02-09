package events

type Action string

const (
	// Transfer actions
	TransferAction = Action("transfer")

	// Contract actions
	DepositAction  = Action("deposit")
	WithdrawAction = Action("withdraw")

	// Swap actions
	BuyAction  = Action("buy")
	SellAction = Action("sell")

	// Liquidity actions
	AddAction    = Action("add")
	RemoveAction = Action("remove")
)
