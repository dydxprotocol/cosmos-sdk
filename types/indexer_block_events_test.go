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
	blockTime := time.Unix(1650000000, 0)
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

//func (s *indexerBlockEventsTestSuite) TestMultipleBlocksIndexerBlockEventManager() {
//	em := sdk.NewIndexerBlockEventManager()
//	txHash := "txHash"
//	txHash1 := "txHash1"
//	newEvent := sdk.NewIndexerTendermintEvent("transfer", "data")
//	newEvent2 := sdk.NewIndexerTendermintEvent("subaccounts", "data2")
//	newEvent3 := sdk.NewIndexerTendermintEvent("order_fill", "data3")
//	blockTime := time.Unix(1650000000, 0)
//	blockHeight := int64(5)
//	em.AddEvent(txHash, newEvent)
//	em.AddEvent(txHash, newEvent3)
//	em.AddEvent(txHash1, newEvent2)
//	em.SetBlockTime(blockTime)
//	em.SetBlockHeight(blockHeight)
//	s.Require().Len(em.ProduceBlock().TxEvents, 2)
//	s.Require().Len(em.ProduceBlock().TxEvents[0].Events, 2)
//	s.Require().Len(em.ProduceBlock().TxEvents[1].Events, 1)
//	s.Require().Equal(em.ProduceBlock().TxEvents[0].Events[0], &newEvent)
//	s.Require().Equal(em.ProduceBlock().TxEvents[0].Events[1], &newEvent3)
//	s.Require().Equal(em.ProduceBlock().TxEvents[1].Events[0], &newEvent2)
//	s.Require().Len(em.ProduceBlock().TxHashes, 2)
//	s.Require().Equal(em.ProduceBlock().TxHashes[0], txHash)
//	s.Require().Equal(em.ProduceBlock().TxHashes[1], txHash1)
//	s.Require().Equal(em.ProduceBlock().Height, uint32(blockHeight))
//	s.Require().Equal(em.ProduceBlock().Time, blockTime)
//}
