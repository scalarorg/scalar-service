package db

func CreateCrossChainDocument(sent ExpectedCrossChainDocument) *CrossChainDocument {
	return &CrossChainDocument{
		ID:          sent.GetID(),
		Type:        sent.GetType(),
		Status:      sent.GetStatus(),
		Source:      sent.GetSource(),
		Destination: sent.GetDestination(),
		CommandID:   sent.GetCommandID(),
	}
}
