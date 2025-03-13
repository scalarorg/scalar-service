package types

type GeneralStats struct {
	TotalTx            uint64 `json:"total_tx"`
	TotalBridgedVolume uint64 `json:"total_bridged_volume"`
	TotalUser          uint64 `json:"total_user"`
}

type AddressAmount struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
}

type ChainAmount struct {
	Chain  string `json:"chain"`
	Amount uint64 `json:"amount"`
}

type PathAmount struct {
	SourceChain      string `json:"source_chain"`
	DestinationChain string `json:"destination_chain"`
	Amount           uint64 `json:"amount"`
}
