package db

type ExpectedCrossChainDocument interface {
	GetID() string
	GetType() CrossChainTx
	GetStatus() string
	GetSource() *SourceDocument
	GetDestination() *DestinationDocument
}
