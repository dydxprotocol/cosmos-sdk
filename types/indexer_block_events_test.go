package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	constants "github.com/cosmos/cosmos-sdk/types/testutil"
	"github.com/stretchr/testify/suite"
	"testing"
)

type indexerBlockEventsTestSuite struct {
	suite.Suite
}

func TestIndexerBlockEventsTestSuite(t *testing.T) {
	suite.Run(t, new(indexerBlockEventsTestSuite))
}

func (s *indexerBlockEventsTestSuite) TestBasicIndexerBlockEventManager() {
	em := sdk.NewIndexerBlockEventManager(constants.BlockHeight, constants.BlockTime)
	em.AddTxnEvent(constants.TxHash, constants.OrderFillSubtype, constants.Data3)
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 1)
	s.Require().Equal(constants.OrderFillEvent, *block.Events[0])
	s.Require().Equal([]string{constants.TxHash}, block.TxHashes)
	s.Require().Equal(constants.BlockHeight, block.Height)
	s.Require().Equal(constants.BlockTime, block.Time)
}

func (s *indexerBlockEventsTestSuite) TestMultipleEventsIndexerBlockEventManager() {
	em := sdk.NewIndexerBlockEventManager(constants.BlockHeight, constants.BlockTime)
	em.AddTxnEvent(constants.TxHash, constants.OrderFillSubtype, constants.Data3)
	em.AddTxnEvent(constants.TxHash, constants.TransferSubtype, constants.Data)
	em.AddTxnEvent(constants.TxHash1, constants.SubaccountSubtype, constants.Data2)
	block := em.ProduceBlock()
	s.Require().Len(block.Events, 3)

	s.Require().Equal(constants.OrderFillEvent, *block.Events[0])
	s.Require().Equal(constants.TransferEvent, *block.Events[1])
	s.Require().Equal(constants.SubaccountEvent, *block.Events[2])
	s.Require().Equal([]string{constants.TxHash, constants.TxHash1}, block.TxHashes)
	s.Require().Equal(constants.BlockHeight, block.Height)
	s.Require().Equal(constants.BlockTime, block.Time)
}

func (s *indexerBlockEventsTestSuite) TestMergeEvents() {
	em := sdk.NewIndexerBlockEventManager(constants.BlockHeight, constants.BlockTime)
	em.AddTxnEvent(constants.TxHash, constants.OrderFillSubtype, constants.Data3)

	em2 := sdk.NewIndexerBlockEventManager(constants.BlockHeight, constants.BlockTime)
	em2.AddTxnEvent(constants.TxHash, constants.TransferSubtype, constants.Data)
	em2.AddTxnEvent(constants.TxHash1, constants.SubaccountSubtype, constants.Data2)
	em.MergeEvents(em2)
	block := em.ProduceBlock()

	s.Require().Len(block.Events, 3)
	s.Require().Equal(constants.OrderFillEvent, *block.Events[0])
	s.Require().Equal(constants.TransferEvent, *block.Events[1])
	s.Require().Equal(constants.SubaccountEvent, *block.Events[2])
	s.Require().Equal([]string{constants.TxHash, constants.TxHash1}, block.TxHashes)
	s.Require().Equal(constants.BlockHeight, block.Height)
	s.Require().Equal(constants.BlockTime, block.Time)
}

//func (s *indexerBlockEventsTestSuite) TestMergeEventFromDifferentBlocks() {
//	em := sdk.NewIndexerBlockEventManager(BlockHeight, BlockTime)
//	em.AddTxnEvent(TxHash, TransferSubtype, Data)
//
//	em2 := sdk.NewIndexerBlockEventManager(uint32(8), BlockTime)
//	em2.AddTxnEvent(TxHash1, SubaccountSubtype, Data2)
//	em.MergeEvents(em2)
//	// Confirm that the block events and transactions are not updated
//	block := em.ProduceBlock()
//	s.Require().Len(block.Events, 1)
//	s.Require().Len(block.TxHashes, 1)
//}
