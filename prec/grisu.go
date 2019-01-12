package prec

import (
	"math"
)

type floatInfo struct {
	mantbits uint
	expbits  uint
	bias     int
}

type decimalSlice struct {
	d      []byte
	nd, dp int
	neg    bool
}

func formatF64(f float64) string {
	return string(genericFtoa(make([]byte, 0, 24), f, 'f', -1, 64))
}

func formatFloat(f float64, fmt byte, prec, bitSize int) string {
	return string(genericFtoa(make([]byte, 0, max(prec+4, 24)), f, fmt, prec, bitSize))
}

func genericFtoa(dst []byte, val float64, fmt byte, prec, bitSize int) []byte {
	var bits uint64
	var flt *floatInfo
	bits = math.Float64bits(val)
	switch bitSize {
	case 32:
		bits = uint64(math.Float32bits(float32(val)))
		flt = &float32info
	case 64:
		bits = math.Float64bits(val)
		flt = &float64info
	default:
		panic("strconv: illegal AppendFloat/FormatFloat bitSize")
	}

	neg := bits>>(flt.expbits+flt.mantbits) != 0
	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1)
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 1<<flt.expbits - 1:
		// Inf, NaN
		var s string
		switch {
		case mant != 0:
			s = "NaN"
		case neg:
			s = "-Inf"
		default:
			s = "+Inf"
		}
		return append(dst, s...)

	case 0:
		// denormalized
		exp++

	default:
		// add implicit top bit
		mant |= uint64(1) << flt.mantbits
	}

	exp += flt.bias

	var digs decimalSlice
	ok := false
	// Try Grisu3 algorithm.
	f := new(extFloat)
	lower, upper := f.AssignComputeBounds(mant, exp, neg, flt)
	var buf [32]byte
	digs.d = buf[:]
	ok = f.ShortestDecimal(&digs, &lower, &upper)
	if !ok {
		return bigFtoa(dst, prec, fmt, neg, mant, exp, flt)
	}
	prec = max(digs.nd-digs.dp, 0)
	return fmtF(dst, neg, digs, prec)
}

func adjustLastDigit(d *decimalSlice, currentDiff, targetDiff, maxDiff, ulpDecimal, ulpBinary uint64) bool {
	if ulpDecimal < 2*ulpBinary {
		// Approximation is too wide.
		return false
	}
	for currentDiff+ulpDecimal/2+ulpBinary < targetDiff {
		d.d[d.nd-1]--
		currentDiff += ulpDecimal
	}
	if currentDiff+ulpDecimal <= targetDiff+ulpDecimal/2+ulpBinary {
		// we have two choices, and don't know what to do.
		return false
	}
	if currentDiff < ulpBinary || currentDiff > maxDiff-ulpBinary {
		// we went too far
		return false
	}
	if d.nd == 1 && d.d[0] == '0' {
		// the number has actually reached zero.
		d.nd = 0
		d.dp = 0
	}
	return true
}


func frexp10Many(a, b, c *extFloat) (exp10 int) {
	exp10, i := c.frexp10()
	a.Multiply(powersOfTen[i])
	b.Multiply(powersOfTen[i])
	return
}
