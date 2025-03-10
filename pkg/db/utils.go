package db

func CreateCrossChainDocument(sent ExpectedCrossChainDocument) *CrossChainDocument {
	return &CrossChainDocument{
		ID:          sent.GetID(),
		Type:        sent.GetType(),
		Status:      string(sent.GetType()),
		Source:      sent.GetSource(),
		Destination: sent.GetDestination(),
	}
}
