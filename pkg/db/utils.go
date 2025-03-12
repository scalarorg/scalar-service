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

func getTimeBucketInterval(bucket string) string {
	switch bucket {
	case "hour":
		return "1 hour"
	case "day":
		return "1 day"
	case "week":
		return "1 week"
	case "month":
		return "1 month"
	default:
		return "1 day" // default interval
	}
}
