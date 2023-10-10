package linkset

func And(a, b []string) (res []string) {
	for i := range a {
		for j := range b {
			if a[i] == b[j] {
				res = append(res, a[i])
				break
			}
		}
	}
	return res
}
func Sub(a, b []string) (res []string) {
	for i := range a {
		var notInB bool = true
		var inRes bool = false
		for j := range b {
			if a[i] == b[j] {
				notInB = false
			}
		}
		for _, v := range res {
			if v == a[i] {
				inRes = true
				break
			}
		}
		if inRes {
			continue
		}

		if notInB {
			res = append(res, a[i])
		}
	}
	return res
}

func Equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
