package main

func filterContains(needle map[string]struct{}, haystack map[string]struct{}) bool {
	if len(haystack) == 0 || len(haystack) < len(needle) {
		return false
	}

	suitable := true
	for v := range needle {
		if _, ok := haystack[v]; !ok {
			suitable = false
			break
		}
	}

	return suitable
}

func filterAny(a map[string]struct{}, b map[string]struct{}) bool {
	low, high := a, b
	if len(a) > len(b) {
		low = b
		high = a
	}

	for v := range low {
		if _, ok := high[v]; ok {
			return true
		}
	}

	return false
}
