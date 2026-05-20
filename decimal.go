package fixnum

import (
	"math"
	"runtime/debug"
	"strconv"
	"strings"

	"go.uber.org/zap"
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
	if d.scale == 0 || o.scale == 0 {
		targetScale := max(o.scale, d.scale)
		if targetScale == 0 {
			targetScale = 1
		}
		zap.S().Errorf("❌ Invalid scale in Decimal.Mul, d: %+v, o: %+v stack:\n%s", d, o, string(debug.Stack()))
		return Decimal{
			v:     0,
			scale: targetScale,
		}
	}

	if mulWillOverflow(d.v, o.v) {
		d = d.loopCompressBy10()
		o = o.loopCompressBy10()
		if mulWillOverflow(d.v, o.v) {
			targetScale := max(o.scale, d.scale)
			zap.S().Errorf("❌ Overflow in Decimal.Mul after compression, d: %+v, o: %+v stack:\n%s", d, o, string(debug.Stack()))
			return Decimal{
				v:     0,
				scale: targetScale,
			}
		}
	}

	targetScale := max(o.scale, d.scale)

	if targetScale == d.scale {
		return Decimal{
			v:     (d.v * o.v) / o.scale,
			scale: d.scale,
		}
	}

	return Decimal{
		v:     (d.v * o.v) / d.scale,
		scale: o.scale,
	}
}

func (d Decimal) Div(o Decimal) Decimal {
	if o.v == 0 {
		zap.S().Errorf("❌ Division by zero in Decimal.Div, d: %+v, o: %+v", d, o)
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

// 取余数
func (d Decimal) Mod(o Decimal) Decimal {
	if o.v == 0 {
		zap.S().Errorf("❌ Mod by zero in Decimal.Mod, d: %+v, o: %+v", d, o)
		return Decimal{v: 0, scale: d.scale}
	}

	targetScale := max(o.scale, d.scale)

	d = d.alignUp(targetScale)
	o = o.alignUp(targetScale)

	return Decimal{
		v:     d.v % o.v,
		scale: targetScale,
	}
}

func (d Decimal) String() string {
	if d.scale == 0 {
		return "0"
	}
	sign := ""
	v := uint64(d.v)
	if d.v < 0 {
		sign = "-"
		v = uint64(-(d.v + 1)) + 1
	}

	scale := uint64(d.scale)
	intPart := v / scale
	fracPart := v % scale

	intStr := sign + strconv.FormatUint(intPart, 10)
	decStr := strconv.FormatUint(fracPart, 10)

	// 补零
	for len(decStr) < len(strconv.FormatInt(d.scale-1, 10)) {
		decStr = "0" + decStr
	}

	// ✅ 去掉右侧多余的 0. 兼容gate等交易所不允许小数末尾有多余的0的情况
	decStr = strings.TrimRight(decStr, "0")

	// ⚠️ 如果小数全是 0，直接返回整数
	if decStr == "" {
		return intStr
	}

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

// 绝对值
func (d Decimal) Abs() Decimal {
	if d.v < 0 {
		return Decimal{
			v:     -d.v,
			scale: d.scale,
		}
	}
	return d
}

// 是否为0
func (d Decimal) IsZero() bool {
	return d.v == 0
}

// ================================[内部函数]====================================

func (d Decimal) alignUp(targetScale int64) Decimal {
	if d.scale == targetScale {
		return d
	}

	if d.scale > targetScale {
		zap.S().Errorf("❌ Precision loss not allowed in Decimal.alignUp, d: %+v, targetScale: %d", d, targetScale)
	}

	if d.scale == 0 {
		zap.S().Errorf("❌ Invalid scale in Decimal.alignUp, d: %+v, targetScale: %d stack:\n%s", d, targetScale, string(debug.Stack()))
		d.scale = 1
	}

	factor := targetScale / d.scale
	return Decimal{
		v:     d.v * factor,
		scale: targetScale,
	}
}

func (d Decimal) compressBy10() (Decimal, bool) {
	if d.v%10 != 0 || d.scale%10 != 0 {
		return d, false
	}

	return Decimal{
		v:     d.v / 10,
		scale: d.scale / 10,
	}, true
}

func (d Decimal) loopCompressBy10() Decimal {
	for {
		compressed, ok := d.compressBy10()
		if !ok {
			return d
		}
		d = compressed
	}
}

func mulWillOverflow(a, b int64) bool {
	if a == 0 || b == 0 {
		return false
	}

	if a > 0 {
		if b > 0 {
			return a > math.MaxInt64/b
		}
		return b < math.MinInt64/a
	}

	if b > 0 {
		return a < math.MinInt64/b
	}

	return b < math.MaxInt64/a
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
