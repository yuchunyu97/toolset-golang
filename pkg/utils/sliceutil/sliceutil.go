package sliceutil

func HasString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, remove func(item string) bool) []string {
	for i := 0; i < len(slice); i++ {
		if remove(slice[i]) {
			slice = append(slice[:i], slice[i+1:]...)
			i--
		}
	}
	return slice
}

func RemoveDuplicate(slice []string) []string {
	set := make(map[string]struct{}, len(slice))
	j := 0
	for _, v := range slice {
		_, ok := set[v]
		if ok {
			continue
		}
		set[v] = struct{}{}
		slice[j] = v
		j++
	}

	return slice[:j]
}
