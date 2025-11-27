package domain

import (
	"strings"
)

func MergeHeaders(collectionHeaders, requestHeaders []KeyValue) []KeyValue {
	if len(collectionHeaders) == 0 {
		return requestHeaders
	}

	// Create a map of request headers by key (case-insensitive) for quick lookup
	requestHeaderMap := make(map[string]KeyValue)
	for _, h := range requestHeaders {
		if h.Enable {
			requestHeaderMap[strings.ToLower(h.Key)] = h
		}
	}

	// Start with collection headers
	merged := make([]KeyValue, 0)

	// Add collection headers that don't have request overrides
	for _, ch := range collectionHeaders {
		if !ch.Enable {
			continue
		}
		keyLower := strings.ToLower(ch.Key)
		if _, hasOverride := requestHeaderMap[keyLower]; !hasOverride {
			merged = append(merged, ch)
		}
	}

	// Add all request headers (they override collection headers)
	for _, rh := range requestHeaders {
		if rh.Enable {
			merged = append(merged, rh)
		}
	}

	return merged
}
