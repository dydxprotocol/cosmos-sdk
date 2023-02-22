package testutil

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
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

var OrderFillEvent = sdk.IndexerTendermintEvent{
	Subtype: OrderFillSubtype,
	Data:    Data3,
	OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
		TransactionIndex: 0,
	},
	EventIndex: 0,
}

var TransferEvent = sdk.IndexerTendermintEvent{
	Subtype: TransferSubtype,
	Data:    Data,
	OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
		TransactionIndex: 0,
	},
	EventIndex: 1,
}

var SubaccountEvent = sdk.IndexerTendermintEvent{
	Subtype: SubaccountSubtype,
	Data:    Data2,
	OrderingWithinBlock: &sdk.IndexerTendermintEvent_TransactionIndex{
		TransactionIndex: 1,
	},
	EventIndex: 0,
}
