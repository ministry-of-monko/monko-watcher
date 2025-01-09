package internal

import (
	"encoding/json"
	"fmt"

	"github.com/algorand/go-algorand-sdk/v2/types"
)

type AssetReport struct {
	Sender   string
	Group    string
	Sent     map[string]map[uint64]uint64
	Received map[string]map[uint64]uint64
	Deltas   map[string]map[uint64]int64
	TxnCount int
}

func NewAssetReport(sender string) *AssetReport {
	return &AssetReport{
		Sender:   sender,
		Sent:     map[string]map[uint64]uint64{},
		Received: map[string]map[uint64]uint64{},
		Deltas:   map[string]map[uint64]int64{},
	}
}

func NewAssetReportWithGroup(sender string, group string) *AssetReport {
	return &AssetReport{
		Sender:   sender,
		Group:    group,
		Sent:     map[string]map[uint64]uint64{},
		Received: map[string]map[uint64]uint64{},
		Deltas:   map[string]map[uint64]int64{},
	}
}

func (report *AssetReport) AddTransaction(tx types.SignedTxnWithAD) {
	switch tx.Txn.Type {
	case types.AssetTransferTx:
		if tx.Txn.AssetAmount > 0 {
			sender := tx.Txn.Sender.String()
			receiver := tx.Txn.AssetReceiver.String()
			assetID := uint64(tx.Txn.XferAsset)
			amount := tx.Txn.AssetAmount

			report.AddSentAmount(sender, assetID, amount)
			report.AddReceivedAmount(receiver, assetID, amount)
		}
	case types.PaymentTx:
		if tx.Txn.Amount > 0 {
			sender := tx.Txn.Sender.String()
			receiver := tx.Txn.Receiver.String()
			amount := uint64(tx.Txn.Amount)

			report.AddSentAmount(sender, 0, amount)
			report.AddReceivedAmount(receiver, 0, amount)
		}
	case types.ApplicationCallTx:
		for _, txn := range tx.EvalDelta.InnerTxns {
			report.AddTransaction(txn)
		}
	}
	report.TxnCount++
}

func (report *AssetReport) CalculateDeltas() {
	for sender, sentAssets := range report.Sent {
		for assetID, amount := range sentAssets {
			if _, ok := report.Deltas[sender]; !ok {
				report.Deltas[sender] = map[uint64]int64{}
			}
			report.Deltas[sender][assetID] -= int64(amount)
		}
	}
	for receiver, receivedAssets := range report.Received {
		for assetID, amount := range receivedAssets {
			if _, ok := report.Deltas[receiver]; !ok {
				report.Deltas[receiver] = map[uint64]int64{}
			}
			report.Deltas[receiver][assetID] += int64(amount)
		}
	}
}

func (report *AssetReport) PrettyPrint() {
	deltaTxn, _ := json.MarshalIndent(report.Deltas, "", "  ")
	senderReport := report.Deltas[report.Sender]
	sendReportString, _ := json.MarshalIndent(senderReport, "", "  ")
	fmt.Printf("********************************\n")
	fmt.Printf("Delta: %s\n", string(deltaTxn))
	fmt.Printf("Sender: %s\n", report.Sender)
	fmt.Printf("Sender Report: %s\n", string(sendReportString))
	fmt.Printf("********************************\n")
}

func (report *AssetReport) AddSentAmount(sender string, assetID uint64, amount uint64) {
	if _, ok := report.Sent[sender]; !ok {
		report.Sent[sender] = map[uint64]uint64{}
	}
	report.Sent[sender][assetID] = amount
}

func (report *AssetReport) AddReceivedAmount(receiver string, assetID uint64, amount uint64) {
	if _, ok := report.Received[receiver]; !ok {
		report.Received[receiver] = map[uint64]uint64{}
	}
	report.Received[receiver][assetID] = amount
}
