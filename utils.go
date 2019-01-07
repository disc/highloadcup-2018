package main

func filterContains(a map[string]struct{}, b map[string]struct{}) bool {
	low, high := a, b
	if len(a) > len(b) {
		low = b
		high = a
	}

	suitable := true
	for v := range low {
		if _, ok := high[v]; !ok {
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
