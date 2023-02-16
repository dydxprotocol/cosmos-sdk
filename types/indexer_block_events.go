package types

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type KafkaBlockEventManager struct {
	block *IndexerTendermintBlock
}

func NewKafkaBlockEventManager() *KafkaBlockEventManager {
	return &KafkaBlockEventManager{&IndexerTendermintBlock{}}
}

func contains[T comparable](s []T, e T) int {
	for i, v := range s {
		if v == e {
			return i
		}
	}
	return -1
}

// AddTransactionHash adds a transaction hash to the block event manager if this is a new transaction
// hash. Returns the index of the new/existing transaction hash in the block event manager.
func (eventManager *KafkaBlockEventManager) AddTransactionHash(txHash string) int {
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

func (eventManager *KafkaBlockEventManager) AddEvent(txHash string, event IndexerTendermintEvent) {
	index := eventManager.AddTransactionHash(txHash)
	if eventManager.block.TxEvents[index].Events == nil {
		eventManager.block.TxEvents[index].Events = make([]*IndexerTendermintEvent, 0)
	}
	eventManager.block.TxEvents[index].Events = append(eventManager.block.TxEvents[index].Events, &event)
}

func (eventManager *KafkaBlockEventManager) SetBlockHeight(blockHeight int64) {
	eventManager.block.Height = uint32(blockHeight)
}

func (eventManager *KafkaBlockEventManager) SetBlockTime(blockTime time.Time) {
	eventManager.block.Time = timestamppb.New(blockTime)
}

func (eventManager *KafkaBlockEventManager) GetBlock() *IndexerTendermintBlock {
	return eventManager.block
}
