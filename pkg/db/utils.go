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

// mergeSortedEntries merges two sorted slices of Entry
func mergeSortedStats[T any](a, b []T, compare func(a, b T) int, merge func(a, b T) T) []T {
	var result []T
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		if compare(a[i], b[j]) == 0 {
			result = append(result, merge(a[i], b[j]))
			i++
			j++
		} else if compare(a[i], b[j]) < 0 {
			result = append(result, a[i])
			i++
		} else {
			result = append(result, b[j])
			j++
		}
	}

	// Append any remaining entries
	for ; i < len(a); i++ {
		result = append(result, a[i])
	}
	for ; j < len(b); j++ {
		result = append(result, b[j])
	}

	return result
}
