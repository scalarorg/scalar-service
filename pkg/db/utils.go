package db

type Options struct {
	Size    int    `json:"size,omitempty"`
	Offset  int    `json:"offset,omitempty"`
	EventId string `json:"event_id,omitempty"`
	TxHash  string `json:"tx_hash,omitempty"`
}
