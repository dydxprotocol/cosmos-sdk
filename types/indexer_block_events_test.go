package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	TxHash            = "txHash"
	TxHash1           = "txHash1"
	TransferSubtype   = "transfer"
	SubaccountSubtype = "subaccount"
	OrderFillSubtype  = "order_fill"
	Data              = "data"
	Data2             = "data2"
	Data3             = "data3"
	BlockHeight       = uint32(5)
)

var BlockTime = time.Unix(1650000000, 0).UTC()

type indexerBlockEventsTestSuite struct {
	suite.Suite
}

func TestIndexerBlockEventsTestSuite(t *testing.T) {
	suite.Run(t, new(indexerBlockEventsTestSuite))
}

func (s *indexerBlockEventsTestSuite) TestBasicIndexerBlockEventManager() {
	em := sdk.NewIndexerBlockEventManager(BlockHeight, BlockTime)
	em.AddTxnEvent(TxHash, TransferSubtype, Data)
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 1)
	expectedEvent := &sdk.IndexerTendermintEvent{
		Subtype: TransferSubtype,
		Data:    Data,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 0,
	}
	s.Require().Equal(*block.Events[0], *expectedEvent)
	s.Require().Len(block.TxHashes, 1)
	s.Require().Equal(block.TxHashes[0], TxHash)
	s.Require().Equal(block.Height, BlockHeight)
	s.Require().Equal(block.Time, BlockTime)
}

func (s *indexerBlockEventsTestSuite) TestMultipleEventsIndexerBlockEventManager() {
	em := sdk.NewIndexerBlockEventManager(BlockHeight, BlockTime)
	em.AddTxnEvent(TxHash, OrderFillSubtype, Data3)
	em.AddTxnEvent(TxHash, TransferSubtype, Data)
	em.AddTxnEvent(TxHash1, SubaccountSubtype, Data2)
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 3)
	expectedOrderFillEvent := sdk.IndexerTendermintEvent{
		Subtype: OrderFillSubtype,
		Data:    Data3,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 0,
	}

	expectedTransferEvent := sdk.IndexerTendermintEvent{
		Subtype: TransferSubtype,
		Data:    Data,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 1,
	}

	expectedSubaccountEvent := sdk.IndexerTendermintEvent{
		Subtype: SubaccountSubtype,
		Data:    Data2,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 1,
		},
		EventIndex: 0,
	}
	s.Require().Equal(*block.Events[0], expectedOrderFillEvent)
	s.Require().Equal(*block.Events[1], expectedTransferEvent)
	s.Require().Equal(*block.Events[2], expectedSubaccountEvent)
	s.Require().Len(block.TxHashes, 2)
	s.Require().Equal(block.TxHashes[0], TxHash)
	s.Require().Equal(block.TxHashes[1], TxHash1)
	s.Require().Equal(block.Height, BlockHeight)
	s.Require().Equal(block.Time, BlockTime)
}

func (s *indexerBlockEventsTestSuite) TestMergeEvents() {
	em := sdk.NewIndexerBlockEventManager(BlockHeight, BlockTime)
	em.AddTxnEvent(TxHash, TransferSubtype, Data)

	em2 := sdk.NewIndexerBlockEventManager(BlockHeight, BlockTime)
	em2.AddTxnEvent(TxHash, SubaccountSubtype, Data2)
	em2.AddTxnEvent(TxHash1, SubaccountSubtype, Data2)
	em.MergeEvents(em2)
	block := em.ProduceBlock()

	s.Require().Len(block.Events, 3)
	s.Require().Equal(sdk.IndexerTendermintEvent{
		Subtype: TransferSubtype,
		Data:    Data,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 0,
	}, *block.Events[0])
	s.Require().Equal(sdk.IndexerTendermintEvent{
		Subtype: SubaccountSubtype,
		Data:    Data2,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 0,
		},
		EventIndex: 1,
	}, *block.Events[1])
	s.Require().Equal(sdk.IndexerTendermintEvent{
		Subtype: SubaccountSubtype,
		Data:    Data2,
		OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
			TransactionIndex: 1,
		},
		EventIndex: 0,
	}, *block.Events[2])
	s.Require().Len(block.TxHashes, 2)
	s.Require().Equal(block.TxHashes[0], TxHash)
	s.Require().Equal(block.TxHashes[1], TxHash1)
	s.Require().Equal(block.Height, BlockHeight)
	s.Require().Equal(block.Time, BlockTime)
}

func (s *indexerBlockEventsTestSuite) TestMergeEventFromDifferentBlocks() {
	em := sdk.NewIndexerBlockEventManager(BlockHeight, BlockTime)
	em.AddTxnEvent(TxHash, TransferSubtype, Data)

	em2 := sdk.NewIndexerBlockEventManager(uint32(8), BlockTime)
	em2.AddTxnEvent(TxHash1, SubaccountSubtype, Data2)
	em.MergeEvents(em2)
	// Confirm that the block events and transactions are not updated
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 1)
	s.Require().Len(block.TxHashes, 1)
}
