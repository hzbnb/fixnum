package decimal

import (
	"strconv"
)

type Decimal struct {
	v     int64
	scale int64
}

func (d Decimal) Add(o Decimal) Decimal {
	targetScale := max(o.scale, d.scale)

	d = d.alignUp(targetScale)
	o = o.alignUp(targetScale)

	return Decimal{
		v:     d.v + o.v,
		scale: targetScale,
	}
}

func (d Decimal) Sub(o Decimal) Decimal {
	targetScale := max(o.scale, d.scale)

	d = d.alignUp(targetScale)
	o = o.alignUp(targetScale)

	return Decimal{
		v:     d.v - o.v,
		scale: targetScale,
	}
}

func (d Decimal) Mul(o Decimal) Decimal {
	// (d.v / d.scale) * (o.v / o.scale)
	// = (d.v * o.v) / (d.scale * o.scale)

	// 我们要保持 d.scale：
	// => (d.v * o.v) / o.scale

	return Decimal{
		v:     (d.v * o.v) / o.scale,
		scale: d.scale,
	}
}

func (d Decimal) Div(o Decimal) Decimal {
	if o.v == 0 {
		// zap.S().Errorf("❌ Division by zero in Decimal.Div, d: %+v, o: %+v", d, o)
		return Decimal{v: 0, scale: d.scale} // 或者 panic
	}

	targetScale := max(o.scale, d.scale)

	d = d.alignUp(targetScale)
	o = o.alignUp(targetScale)

	return Decimal{
		v:     (d.v * targetScale) / o.v,
		scale: targetScale,
	}
}

func (d Decimal) String() string {
	if d.scale == 0 {
		return "0"
	}
	intPart := d.v / d.scale
	fracPart := d.v % d.scale

	intStr := strconv.FormatInt(intPart, 10)
	decStr := strconv.FormatInt(fracPart, 10)

	// 补零
	for len(decStr) < len(strconv.FormatInt(d.scale-1, 10)) {
		decStr = "0" + decStr
	}

	// // ✅ 去掉右侧多余的 0
	// decStr = strings.TrimRight(decStr, "0")

	// // ⚠️ 如果小数全是 0，直接返回整数
	// if decStr == "" {
	// 	return intStr
	// }

	return intStr + "." + decStr
}

// 简单计算比较用。不能传参数
func (d Decimal) Float64() float64 {
	if d.scale == 0 {
		return 0
	}
	return float64(d.v) / float64(d.scale)
}

func (d Decimal) Int64() int64 {
	if d.scale == 0 {
		return 0
	}
	return d.v / d.scale
}

func (d Decimal) GetScale() int64 {
	return d.scale
}

// 等于
func (d Decimal) Eq(o Decimal) bool {
	return d.cmp(o) == 0
}

// 大于
func (d Decimal) Gt(o Decimal) bool {
	return d.cmp(o) == 1
}

// 大于等于
func (d Decimal) Gte(o Decimal) bool {
	return d.cmp(o) >= 0
}

// 小于
func (d Decimal) Lt(o Decimal) bool {
	return d.cmp(o) == -1
}

// 小于等于
func (d Decimal) Lte(o Decimal) bool {
	return d.cmp(o) <= 0
}

// ================================[内部函数]====================================

func (d Decimal) alignUp(targetScale int64) Decimal {
	if d.scale == targetScale {
		return d
	}

	if d.scale > targetScale {
		// zap.S().Errorf("❌ Precision loss not allowed in Decimal.alignUp, d: %+v, targetScale: %d", d, targetScale)
	}

	if d.scale == 0 {
		// zap.S().Errorf("❌ Invalid scale in Decimal.alignUp, d: %+v, targetScale: %d", d, targetScale)
		d.scale = 1
	}

	factor := targetScale / d.scale
	return Decimal{
		v:     d.v * factor,
		scale: targetScale,
	}
}

func (d Decimal) cmp(o Decimal) int {
	targetScale := max(o.scale, d.scale)

	d = d.alignUp(targetScale)
	o = o.alignUp(targetScale)

	switch {
	case d.v > o.v:
		return 1
	case d.v < o.v:
		return -1
	default:
		return 0
	}
}
