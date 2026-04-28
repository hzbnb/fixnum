package fixnum

func GetDigits(s string) int64 {
	dot := -1

	for i, c := range s {
		if c == '.' {
			dot = i
			break
		}
	}

	if dot == -1 {
		return 0
	}

	frac := s[dot+1:]

	// 去掉尾部 0
	end := len(frac)
	for end > 0 && frac[end-1] == '0' {
		end--
	}

	return int64(end)
}

func GetScaleFromDigits(digits int64) int64 {
	scale := int64(1)
	for range digits {
		scale *= 10
	}
	return scale
}
