package fixnum

import "math"

func NewFromString(s string, digits int64) Decimal {
	var scale int64 = 1
	for range digits {
		scale *= 10
	}

	var intPart int64 = 0
	var fracPart int64 = 0
	var fracLen int64 = 0
	dot := false
	neg := false

	for i, c := range s {
		if i == 0 && c == '-' {
			neg = true
			continue
		}
		if c == '.' {
			dot = true
			continue
		}
		if c < '0' || c > '9' {
			continue // 可选：忽略非法字符 or 直接 panic
		}
		if !dot {
			intPart = intPart*10 + int64(c-'0')
		} else {
			fracPart = fracPart*10 + int64(c-'0')
			fracLen++
		}
	}

	// 截断 or 补齐
	if fracLen > digits {
		// 向下截断（交易系统标准）
		for i := int64(0); i < fracLen-digits; i++ {
			fracPart /= 10
		}
	} else {
		for fracLen < digits {
			fracPart *= 10
			fracLen++
		}
	}

	v := intPart*scale + fracPart
	if neg {
		v = -v
	}

	return Decimal{
		v:     v,
		scale: scale,
	}
}

func NewFromInt(i int64) Decimal {
	var scale int64 = 1

	return Decimal{
		v:     i * scale,
		scale: scale,
	}
}

func NewFromDecimalWithScale(d Decimal, scale int64) Decimal {
	if scale <= 0 {
		return d
	}

	if d.scale == 0 {
		return Decimal{v: 0, scale: scale}
	}

	if scale == d.scale {
		return d
	}

	if scale > d.scale {
		factor := scale / d.scale
		return Decimal{
			v:     d.v * factor,
			scale: scale,
		}
	}

	factor := d.scale / scale
	return Decimal{
		v:     d.v / factor,
		scale: scale,
	}
}

func NewDecimalFromFloat64(f float64, digits int64) Decimal {
	// 计算 scale
	scale := int64(1)
	for range digits {
		scale *= 10
	}

	// 关键：放大 + 截断（floor）
	v := int64(f * float64(scale))

	return Decimal{
		v:     v,
		scale: scale,
	}
}

func NewZero() Decimal {
	return Decimal{v: 0, scale: 1}
}

func NewFromStringWithScale(s string, scale int64) Decimal {
	scaleCopy := scale

	var digits int64 = 0
	for scaleCopy > 1 {
		scaleCopy /= 10
		digits++
	}

	var intPart int64 = 0
	var fracPart int64 = 0
	var fracLen int64 = 0
	dot := false

	for _, c := range s {
		if c == '.' {
			dot = true
			continue
		}
		if c < '0' || c > '9' {
			continue // 可选：忽略非法字符 or 直接 panic
		}
		if !dot {
			intPart = intPart*10 + int64(c-'0')
		} else {
			fracPart = fracPart*10 + int64(c-'0')
			fracLen++
		}
	}

	// 截断 or 补齐
	if fracLen > digits {
		// 向下截断（交易系统标准）
		for i := int64(0); i < fracLen-digits; i++ {
			fracPart /= 10
		}
	} else {
		for fracLen < digits {
			fracPart *= 10
			fracLen++
		}
	}

	v := intPart*scale + fracPart

	return Decimal{
		v:     v,
		scale: scale,
	}
}

func NewFromFloat64WithScale(f float64, scale int64) Decimal {
	v := int64(math.Round(f * float64(scale)))

	return Decimal{
		v:     v,
		scale: scale,
	}
}
