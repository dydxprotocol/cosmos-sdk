package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type indexerBlockEventsTestSuite struct {
	suite.Suite
}

func TestIndexerBlockEventsTestSuite(t *testing.T) {
	suite.Run(t, new(indexerBlockEventsTestSuite))
}

func (s *indexerBlockEventsTestSuite) TestBasicIndexerBlockEventManager() {
	em := sdk.NewIndexerBlockEventManager()
	txHash := "txHash"
	subType := "transfer"
	data := "data"
	blockTime := time.Unix(1650000000, 0).UTC()
	blockHeight := int64(5)
	em.AddTxnEvent(txHash, subType, data)
	em.SetBlockTime(blockTime)
	em.SetBlockHeight(blockHeight)
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 1)
	expectedEvent := &sdk.IndexerTendermintEvent{
		Subtype: subType,
		Data:    data,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 0,
	}
	s.Require().Equal(*block.Events[0], *expectedEvent)
	s.Require().Len(block.TxHashes, 1)
	s.Require().Equal(block.TxHashes[0], txHash)
	s.Require().Equal(block.Height, uint32(blockHeight))
	s.Require().Equal(block.Time, blockTime)
}

func (s *indexerBlockEventsTestSuite) TestMultipleEventsIndexerBlockEventManager() {
	em := sdk.NewIndexerBlockEventManager()
	txHash := "txHash"
	txHash1 := "txHash1"
	blockTime := time.Unix(1650000000, 0).UTC()
	blockHeight := int64(5)
	em.AddTxnEvent(txHash, "order_fill", "data3")
	em.AddTxnEvent(txHash, "transfer", "data")
	em.AddTxnEvent(txHash1, "subaccounts", "data2")
	em.SetBlockTime(blockTime)
	em.SetBlockHeight(blockHeight)
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 3)
	expectedOrderFillEvent := sdk.IndexerTendermintEvent{
		Subtype: "order_fill",
		Data:    "data3",
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 0,
	}

	expectedTransferEvent := sdk.IndexerTendermintEvent{
		Subtype: "transfer",
		Data:    "data",
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 1,
	}

	expectedSubaccountEvent := sdk.IndexerTendermintEvent{
		Subtype: "subaccounts",
		Data:    "data2",
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 1,
		},
		EventIndex: 0,
	}
	s.Require().Equal(*block.Events[0], expectedOrderFillEvent)
	s.Require().Equal(*block.Events[1], expectedTransferEvent)
	s.Require().Equal(*block.Events[2], expectedSubaccountEvent)
	s.Require().Len(block.TxHashes, 2)
	s.Require().Equal(block.TxHashes[0], txHash)
	s.Require().Equal(block.TxHashes[1], txHash1)
	s.Require().Equal(block.Height, uint32(blockHeight))
	s.Require().Equal(block.Time, blockTime)
}

func (s *indexerBlockEventsTestSuite) TestMergeEvents() {
	em := sdk.NewIndexerBlockEventManager()
	txHash := "txHash"
	subType := "transfer"
	data := "data"
	blockTime := time.Unix(1650000000, 0).UTC()
	blockHeight := int64(5)
	em.SetBlockTime(blockTime)
	em.SetBlockHeight(blockHeight)
	em.AddTxnEvent(txHash, subType, data)

	em2 := sdk.NewIndexerBlockEventManager()
	txHash2 := "txHash2"
	subType2 := "subaccount"
	data2 := "data2"
	em2.SetBlockTime(blockTime)
	em2.SetBlockHeight(blockHeight)
	em2.AddTxnEvent(txHash, subType2, data2)
	em2.AddTxnEvent(txHash2, subType2, data2)
	em.MergeEvents(em2)
	block := em.ProduceBlock()

	s.Require().Len(block.Events, 3)
	s.Require().Equal(sdk.IndexerTendermintEvent{
		Subtype: subType,
		Data:    data,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 0,
	}, *block.Events[0])
	s.Require().Equal(sdk.IndexerTendermintEvent{
		Subtype: subType2,
		Data:    data2,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 1,
	}, *block.Events[1])
	s.Require().Equal(sdk.IndexerTendermintEvent{
		Subtype: subType2,
		Data:    data2,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 1,
		},
		EventIndex: 0,
	}, *block.Events[2])
	s.Require().Len(block.TxHashes, 2)
	s.Require().Equal(block.TxHashes[0], txHash)
	s.Require().Equal(block.Height, uint32(blockHeight))
	s.Require().Equal(block.Time, blockTime)
}

func (s *indexerBlockEventsTestSuite) TestMergeEventFromDifferentBlocks() {
	em := sdk.NewIndexerBlockEventManager()
	blockTime := time.Unix(1650000000, 0).UTC()
	blockHeight := int64(5)
	em.SetBlockTime(blockTime)
	em.SetBlockHeight(blockHeight)
	em.AddTxnEvent("txHash", "subType", "data")

	em2 := sdk.NewIndexerBlockEventManager()
	blockHeight2 := int64(8)
	em2.SetBlockTime(blockTime)
	em2.SetBlockHeight(blockHeight2)
	em2.AddTxnEvent("txHash2", "subType", "data")
	em.MergeEvents(em2)
	// Confirm that the block events and transactions are not updated
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 1)
	s.Require().Len(block.TxHashes, 1)
}
