package types

import (
	"time"
)

type IndexerBlockEventManager struct {
	block *IndexerTendermintBlock
}

func NewIndexerTendermintEvent(subType string, data string) IndexerTendermintEvent {
	return IndexerTendermintEvent{
		Subtype: subType,
		Data:    data,
	}
}

// NewIndexerBlockEventManager returns a new IndexerBlockEventManager.
// This should be called in BeginBlocker.
func NewIndexerBlockEventManager() *IndexerBlockEventManager {
	return &IndexerBlockEventManager{&IndexerTendermintBlock{}}
}

func contains[T comparable](s []T, e T) int {
	for i, v := range s {
		if v == e {
			return i
		}
	}
	return -1
}

func (eventManager *IndexerBlockEventManager) SetBlock(block *IndexerTendermintBlock) {
	eventManager.block = block
}

// addTransactionHash adds a transaction hash to the block event manager if this is a new transaction
// hash. Returns the index of the new/existing transaction hash in the block event manager.
func (eventManager *IndexerBlockEventManager) addTransactionHash(txHash string) int {
	if eventManager.block.TxHashes == nil {
		eventManager.block.TxHashes = []string{}
	}
	if index := contains(eventManager.block.TxHashes, txHash); index != -1 {
		return index
	}
	eventManager.block.TxHashes = append(eventManager.block.TxHashes, txHash)
	eventManager.block.TxEvents = append(eventManager.block.TxEvents, &TransactionEvents{})
	return len(eventManager.block.TxEvents) - 1
}

// AddEvent adds an event to the block event manager. If the transaction hash is not already in the
// block event manager, it is also added.
func (eventManager *IndexerBlockEventManager) AddEvent(txHash string, event IndexerTendermintEvent) {
	index := eventManager.addTransactionHash(txHash)
	if eventManager.block.TxEvents[index].Events == nil {
		eventManager.block.TxEvents[index].Events = make([]*IndexerTendermintEvent, 0)
	}
	eventManager.block.TxEvents[index].Events = append(eventManager.block.TxEvents[index].Events, &event)
}

// SetBlockHeight sets the block height of the block event manager.
func (eventManager *IndexerBlockEventManager) SetBlockHeight(blockHeight int64) {
	eventManager.block.Height = uint32(blockHeight)
}

// SetBlockTime sets the block time of the block event manager.
func (eventManager *IndexerBlockEventManager) SetBlockTime(blockTime time.Time) {
	eventManager.block.Time = blockTime
}

// GetBlock returns the block. It should only be called in EndBlocker. Otherwise, the block is
// incomplete.
func (eventManager *IndexerBlockEventManager) GetBlock() *IndexerTendermintBlock {
	return eventManager.block
}
