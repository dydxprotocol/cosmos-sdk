package types

import (
	"time"
)

type IndexerBlockEventManager struct {
	height   uint32
	time     time.Time
	txHashes []string
	// map from tx hash to list of tx events
	txEventsMap map[string][]IndexerTendermintEvent
}

// NewIndexerBlockEventManager returns a new IndexerBlockEventManager.
// This should be called in BeginBlocker.
func NewIndexerBlockEventManager() *IndexerBlockEventManager {
	return &IndexerBlockEventManager{
		txHashes:    []string{},
		txEventsMap: make(map[string][]IndexerTendermintEvent),
	}
}

func (eventManager *IndexerBlockEventManager) SetAttributes(mgr *IndexerBlockEventManager) {
	eventManager.height = mgr.height
	eventManager.time = mgr.time
	// deep copy txHashes
	eventManager.txHashes = make([]string, len(mgr.txHashes))
	eventManager.txHashes = append(eventManager.txHashes, mgr.txHashes...)
	// deep copy txEventsMap
	eventManager.txEventsMap = make(map[string][]IndexerTendermintEvent)
	for txHash, events := range mgr.txEventsMap {
		eventManager.txEventsMap[txHash] = make([]IndexerTendermintEvent, len(events))
		eventManager.txEventsMap[txHash] = append(eventManager.txEventsMap[txHash], events...)
	}
}

// AddTxnEvent adds a transaction event to the block event manager. If the transaction hash is not already in the
// block event manager, it is also added.
func (eventManager *IndexerBlockEventManager) AddTxnEvent(txHash string, subType string, data string) {
	event := IndexerTendermintEvent{
		Subtype: subType,
		Data:    data,
	}
	if txEvents, ok := eventManager.txEventsMap[txHash]; ok {
		eventManager.txEventsMap[txHash] = append(txEvents, event)
	} else {
		eventManager.txHashes = append(eventManager.txHashes, txHash)
		eventManager.txEventsMap[txHash] = []IndexerTendermintEvent{event}
	}
}

// SetBlockHeight sets the block height of the block event manager.
func (eventManager *IndexerBlockEventManager) SetBlockHeight(blockHeight int64) {
	eventManager.height = uint32(blockHeight)
}

// SetBlockTime sets the block time of the block event manager.
func (eventManager *IndexerBlockEventManager) SetBlockTime(blockTime time.Time) {
	eventManager.time = blockTime
}

// ProduceBlock returns the block. It should only be called in EndBlocker. Otherwise, the block is
// incomplete.
func (eventManager *IndexerBlockEventManager) ProduceBlock() *IndexerTendermintBlock {
	// create map from txHash to index
	txHashesMap := make(map[string]int)
	for i, txHash := range eventManager.txHashes {
		txHashesMap[txHash] = i
	}
	// iterate through txEventsMap and add transaction/event indices to each event
	for txHash, events := range eventManager.txEventsMap {
		for i, event := range events {
			event.OrderingWithinBlock = &IndexerTendermintEvent_TransactionIndex{
				TransactionIndex: uint32(txHashesMap[txHash]),
			}
			event.EventIndex = uint32(i)
			events[i] = event
		}
		eventManager.txEventsMap[txHash] = events
	}
	// build list of tx events
	var txEvents []*IndexerTendermintEvent
	for _, txHash := range eventManager.txHashes {
		for _, event := range eventManager.txEventsMap[txHash] {
			txEvents = append(txEvents, &event)
		}
	}
	return &IndexerTendermintBlock{
		Height:   eventManager.height,
		Time:     eventManager.time,
		Events:   txEvents,
		TxHashes: eventManager.txHashes,
	}
}
