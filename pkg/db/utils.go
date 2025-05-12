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

func validateTimeBucketInterval(bucket string) bool {
	if bucket == "microseconds" ||
		bucket == "milliseconds" ||
		bucket == "second" ||
		bucket == "minute" ||
		bucket == "hour" ||
		bucket == "day" ||
		bucket == "week" ||
		bucket == "month" ||
		bucket == "quarter" ||
		bucket == "year" ||
		bucket == "decade" ||
		bucket == "century" ||
		bucket == "millennium" {
		return true
	}
	return false
}
