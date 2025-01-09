package internal

import (
	"encoding/base64"

	"github.com/algorand/go-algorand-sdk/v2/types"
)

type BlockAnalyzer struct {
	assetList map[uint64]struct{}
}

func NewBlockAnalyzer(assetsToWatch ...uint64) *BlockAnalyzer {
	analyzer := &BlockAnalyzer{
		assetList: map[uint64]struct{}{},
	}

	for _, assetID := range assetsToWatch {
		analyzer.assetList[assetID] = struct{}{}
	}

	return analyzer
}

func (analyzer *BlockAnalyzer) AnalyzeBlock(block types.Block) []*AssetReport {
	grouped := map[string][]types.SignedTxnInBlock{}
	singleTxns := []types.SignedTxnInBlock{}
	monkoGroups := map[string]struct{}{}
	reports := []*AssetReport{}

	for _, tx := range block.Payset {

		groupString := base64.StdEncoding.EncodeToString(tx.Txn.Group[:])

		if groupString == "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=" && tx.Txn.Type == types.AssetTransferTx {
			_, shouldTrack := analyzer.assetList[uint64(tx.Txn.XferAsset)]
			if shouldTrack && tx.Txn.AssetAmount > 0 {
				singleTxns = append(singleTxns, tx)
			}
			continue
		}

		grouped[groupString] = append(grouped[groupString], tx)
		shouldTrack := false

		switch tx.Txn.Type {
		case types.AssetTransferTx:
			_, shouldTrack = analyzer.assetList[uint64(tx.Txn.XferAsset)]
		case types.ApplicationCallTx:
			for _, foreign := range tx.Txn.ForeignAssets {
				_, shouldTrack = analyzer.assetList[uint64(foreign)]
				if shouldTrack {
					break
				}
			}
		}

		if shouldTrack {
			monkoGroups[groupString] = struct{}{}
		}
	}

	for _, tx := range singleTxns {
		report := NewAssetReport(tx.Txn.Sender.String())
		report.AddTransaction(tx.SignedTxnWithAD)
		reports = append(reports, report)
	}

	for groupString, group := range grouped {
		if _, ok := monkoGroups[groupString]; ok {
			sender := group[0].Txn.Sender.String()
			report := NewAssetReportWithGroup(sender, groupString)
			for _, tx := range group {
				report.AddTransaction(tx.SignedTxnWithAD)
			}
			reports = append(reports, report)
		}
	}

	return reports
}
