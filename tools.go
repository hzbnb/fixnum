package fixnum

import "go.uber.org/zap"

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

func Min(a, b Decimal) Decimal {
	if a.Lte(b) {
		return a
	}
	return b
}

func Max(a, b Decimal) Decimal {
	if a.Gte(b) {
		return a
	}
	return b
}

// 向下取整到步长的整数倍
func FloorToStep(value Decimal, tick Decimal) Decimal {
	if tick.IsZero() {
		zap.S().Errorf("❌ FloorToStep tick is zero, value: %+v, tick: %+v", value, tick)
		return value
	}

	return value.Sub(value.Mod(tick))
}
