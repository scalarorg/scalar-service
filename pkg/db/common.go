package db

import "time"

type TransferType string

const (
	TransferTypeSendToken      TransferType = "send_token"
	TransferTypeDepositAddress TransferType = "deposit_address"
)

type VoteType string

const (
	VoteTypeVote VoteType = "vote"
)

type VoteEvent string

const (
	VoteEventTokenSent VoteEvent = "token_sent"
)

type ConfirmType string

const (
	ConfirmTypeVote ConfirmType = "vote"
)

// Transfer represents the main transfer record
type TransferDocument struct {
	ID               string       `json:"id"`
	Type             TransferType `json:"type"`
	Status           string       `json:"status"`
	SimplifiedStatus string       `json:"simplified_status"`
	TransferID       uint         `json:"transfer_id"`

	// Relationships
	Send      SendInfo    `json:"send"`
	TimeSpent TimeSpent   `json:"time_spent"`
	Confirm   ConfirmInfo `json:"confirm"`
	Link      LinkInfo    `json:"link"`
	Command   CommandInfo `json:"command"`
	Vote      VoteInfo    `json:"vote"`
}

// TimeInfo represents the timestamp information structure
type TimeInfo struct {
	MS      int64 `json:"ms"`
	Hour    int64 `json:"hour"`
	Day     int64 `json:"day"`
	Week    int64 `json:"week"`
	Month   int64 `json:"month"`
	Quarter int64 `json:"quarter"`
	Year    int64 `json:"year"`
}

func FormatTimeInfo(timestamp time.Time) TimeInfo {
	// Keep MS as is since it's the base unit
	ms := timestamp.UnixNano() / int64(time.Millisecond)

	// Truncate to start of each period
	hourStart := timestamp.Truncate(time.Hour)
	dayStart := timestamp.Truncate(24 * time.Hour)
	// Week starts on Monday by default
	weekStart := timestamp.Truncate(24 * time.Hour * 7)
	// For month, we need to set to first day
	monthStart := time.Date(timestamp.Year(), timestamp.Month(), 1, 0, 0, 0, 0, timestamp.Location())
	// Quarter - set to start of the quarter
	quarter := (timestamp.Month()-1)/3*3 + 1
	quarterStart := time.Date(timestamp.Year(), quarter, 1, 0, 0, 0, 0, timestamp.Location())
	// Year - set to January 1st
	yearStart := time.Date(timestamp.Year(), 1, 1, 0, 0, 0, 0, timestamp.Location())

	return TimeInfo{
		MS:      ms,
		Hour:    hourStart.UnixNano() / int64(time.Millisecond),
		Day:     dayStart.UnixNano() / int64(time.Millisecond),
		Week:    weekStart.UnixNano() / int64(time.Millisecond),
		Month:   monthStart.UnixNano() / int64(time.Millisecond),
		Quarter: quarterStart.UnixNano() / int64(time.Millisecond),
		Year:    yearStart.UnixNano() / int64(time.Millisecond),
	}
}

// SendInfo represents the send information
type SendInfo struct {
	TxHash                   string   `json:"txhash"`
	Height                   uint64   `json:"height"`
	Status                   string   `json:"status"`
	Type                     string   `json:"type"`
	CreatedAt                TimeInfo `json:"created_at"`
	SourceChain              string   `json:"source_chain"`
	SenderAddress            string   `json:"sender_address"`
	RecipientAddress         string   `json:"recipient_address"`
	Denom                    string   `json:"denom"`
	Amount                   float64  `json:"amount"`
	Value                    float64  `json:"value"`
	DestinationChain         string   `json:"destination_chain"`
	OriginalSourceChain      string   `json:"original_source_chain"`
	Fee                      float64  `json:"fee"`
	FeeValue                 float64  `json:"fee_value"`
	AmountReceived           float64  `json:"amount_received"`
	OriginalDestinationChain string   `json:"original_destination_chain"`
	InsufficientFee          bool     `json:"insufficient_fee"`
}

// TimeSpent represents the time spent information
type TimeSpent struct {
	SendConfirm int `json:"send_confirm"`
	ConfirmVote int `json:"confirm_vote"`
	Total       int `json:"total"`
	ConfirmIBC  int `json:"confirm_ibc"`
	VoteIBC     int `json:"vote_ibc"`
}

// ConfirmInfo represents the confirm information
type ConfirmInfo struct {
	Amount           string      `json:"amount"`
	SourceChain      string      `json:"source_chain"`
	DepositAddress   string      `json:"deposit_address"`
	CreatedAt        TimeInfo    `json:"created_at"`
	DestinationChain string      `json:"destination_chain"`
	Denom            string      `json:"denom"`
	TransferID       uint        `json:"transfer_id"`
	Type             ConfirmType `json:"type"`
	TxHash           string      `json:"txhash"`
	Height           uint64      `json:"height"`
	Status           string      `json:"status"`
}

// LinkInfo represents the link information
type LinkInfo struct {
	OriginalSourceChain      string   `json:"original_source_chain"`
	DepositAddress           string   `json:"deposit_address"`
	CreatedAt                TimeInfo `json:"created_at"`
	DestinationChain         string   `json:"destination_chain"`
	RecipientAddress         string   `json:"recipient_address"`
	Type                     string   `json:"type"`
	TxHash                   string   `json:"txhash"`
	SenderAddress            string   `json:"sender_address"`
	SourceChain              string   `json:"source_chain"`
	ID                       string   `json:"id"`
	Denom                    string   `json:"denom"`
	OriginalDestinationChain string   `json:"original_destination_chain"`
	Height                   uint64   `json:"height"`
}

// CommandInfo represents the command information
type CommandInfo struct {
	Chain            string   `json:"chain"`
	CommandID        string   `json:"command_id"`
	LogIndex         uint     `json:"log_index"`
	BatchID          string   `json:"batch_id"`
	BlockNumber      uint64   `json:"block_number"`
	BlockTimestamp   int64    `json:"block_timestamp"`
	CreatedAt        TimeInfo `json:"created_at"`
	Executed         bool     `json:"executed"`
	TransactionIndex uint     `json:"transaction_index"`
	TransferID       uint     `json:"transfer_id"`
	TransactionHash  string   `json:"transaction_hash"`
}

type VoteInfo struct {
	TransactionID    string    `json:"transaction_id"`
	PollID           string    `json:"poll_id"`
	SourceChain      string    `json:"source_chain"`
	CreatedAt        TimeInfo  `json:"created_at"`
	DestinationChain string    `json:"destination_chain"`
	Confirmation     bool      `json:"confirmation"`
	Type             VoteType  `json:"type"`
	Event            VoteEvent `json:"event"`
	TxHash           string    `json:"txhash"`
	Height           uint64    `json:"height"`
	Status           string    `json:"status"`
	TransferID       uint      `json:"transfer_id"`
	Success          bool      `json:"success"`
}
