package prec

import (
	"math"
)

// 456.1929434 -> 7
func floatPrec(val float64) int {
	//dst := make([]byte, 0, 24)
	var bits uint64
	var flt *floatInfo
	bits = math.Float64bits(val)
	bits = math.Float64bits(val)
	flt = &float64info

	neg := bits>>(flt.expbits+flt.mantbits) != 0
	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1)
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 1<<flt.expbits - 1:
		// Inf, NaN
		//var s string
		//switch {
		//case mant != 0:
		//	s = "NaN"
		//case neg:
		//	s = "-Inf"
		//default:
		//	s = "+Inf"
		//}
		return 0

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
		d := new(decimal)
		d.Assign(mant)
		d.Shift(exp - int(flt.mantbits))
		var digs decimalSlice
		roundShortest(d, mant, exp, flt)
		digs = decimalSlice{d: d.d[:], nd: d.nd, dp: d.dp}
		// Precision for shortest representation mode.
		return max(digs.nd-digs.dp, 0)
	}
	return max(digs.nd-digs.dp, 0)
}

// false [2 5 1] [5 3 6] -> 152.536 true
func bytes2Float64(neg bool, integer []uint8, fraction []uint8) (f float64, ok bool) {
	fl := len(fraction)
	il := len(integer)
	exp := -fl
	var mantiss = uint64(0)
	for i := il - 1; i >= 0; i-- {
		mantiss *= 10
		mantiss += uint64(integer[i])
	}
	for i := 0; i < fl; i++ {
		mantiss *= 10
		mantiss += uint64(fraction[i])
	}

	if mantiss>>float64info.mantbits != 0 {
		return
	}
	f = float64(mantiss)
	if neg {
		f = -f
	}
	switch {
	case exp == 0:
		// an integer.
		return f, true
	// Exact integers are <= 10^15.
	// Exact powers of ten are <= 10^22.
	case exp > 0 && exp <= 15+22: // int * 10^k
		// If exponent is big but number of digits is not,
		// can move a few zeros into the integer part.
		if exp > 22 {
			f *= float64pow10[exp-22]
			exp = 22
		}
		if f > 1e15 || f < -1e15 {
			// the exponent was really too large.
			return
		}
		return f * float64pow10[exp], true
	case exp < 0 && exp >= -22: // int / 10^k
		return f / float64pow10[-exp], true
	}
	return
}

// 493.483 -> [3 9 4] [4 8 3] 整部部分倒叙
func float642Bytes(val float64) (bool, []byte, []byte) {
	//dst := make([]byte, 0, 24)
	var bits uint64
	var flt *floatInfo
	bits = math.Float64bits(val)
	flt = &float64info

	neg := bits>>(flt.expbits+flt.mantbits) != 0
	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1)
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 1<<flt.expbits - 1:
		return neg, nil, nil
	case 0:
		// denormalized
		exp++
	default:
		// add implicit top bit
		mant |= uint64(1) << flt.mantbits
	}

	exp += flt.bias

	var prec int
	var digs decimalSlice
	ok := false
	// Try Grisu3 algorithm.
	f := new(extFloat)
	lower, upper := f.AssignComputeBounds(mant, exp, neg, flt)
	var buf [32]byte
	digs.d = buf[:]
	ok = f.ShortestDecimal(&digs, &lower, &upper)
	if !ok {
		d := new(decimal)
		d.Assign(mant)
		d.Shift(exp - int(flt.mantbits))
		var digs decimalSlice
		roundShortest(d, mant, exp, flt)
		digs = decimalSlice{d: d.d[:], nd: d.nd, dp: d.dp}
		// Precision for shortest representation mode.
		prec = max(digs.nd-digs.dp, 0)
	} else {
		prec = max(digs.nd-digs.dp, 0)
	}
	//
	var integer, fraction []byte

	// integer, padded with zeros as needed.
	if digs.dp > 0 {
		m := min(digs.nd, digs.dp)
		for ; m < digs.dp; m++ {
			integer = append(integer, 0)
		}
		for i := m-1; i >= 0; i-- {
			integer = append(integer, digs.d[i]-'0')
		}
	} else {
		integer = append(integer, 0)
	}

	// fraction
	if prec > 0 {
		for i := 0; i < prec; i++ {
			ch := byte(0)
			if j := digs.dp + i; 0 <= j && j < digs.nd {
				ch = digs.d[j]-'0'
			}
			fraction = append(fraction, ch)
		}
	}

	return neg, integer, fraction
}

// 284.491 -> false 3 284 491
func float642Uints(val float64) (bool, int, uint64, uint64) {
	//dst := make([]byte, 0, 24)
	var bits uint64
	var flt *floatInfo
	bits = math.Float64bits(val)
	flt = &float64info

	neg := bits>>(flt.expbits+flt.mantbits) != 0
	exp := int(bits>>flt.mantbits) & (1<<flt.expbits - 1)
	mant := bits & (uint64(1)<<flt.mantbits - 1)

	switch exp {
	case 1<<flt.expbits - 1:
		return neg, 0, 0, 0
	case 0:
		// denormalized
		exp++
	default:
		// add implicit top bit
		mant |= uint64(1) << flt.mantbits
	}

	exp += flt.bias

	var prec int
	var digs decimalSlice
	ok := false
	// Try Grisu3 algorithm.
	f := new(extFloat)
	lower, upper := f.AssignComputeBounds(mant, exp, neg, flt)
	var buf [32]byte
	digs.d = buf[:]
	ok = f.ShortestDecimal(&digs, &lower, &upper)
	if !ok {
		d := new(decimal)
		d.Assign(mant)
		d.Shift(exp - int(flt.mantbits))
		var digs decimalSlice
		roundShortest(d, mant, exp, flt)
		digs = decimalSlice{d: d.d[:], nd: d.nd, dp: d.dp}
		// Precision for shortest representation mode.
		prec = max(digs.nd-digs.dp, 0)
	} else {
		prec = max(digs.nd-digs.dp, 0)
	}
	//
	var integer, fraction uint64

	// integer, padded with zeros as needed.
	if digs.dp > 0 {
		m := min(digs.nd, digs.dp)
		for i := 0; i < m; i++ {
			integer *= 10
			integer += uint64(digs.d[i]-'0')
		}
		for ; m < digs.dp; m++ {
			integer *= 10
		}
	}

	// fraction
	if prec > 0 {
		for i := 0; i < prec; i++ {
			ch := uint64(0)
			if j := digs.dp + i; 0 <= j && j < digs.nd {
				ch = uint64(digs.d[j]-'0')
			}
			fraction *= 10
			fraction += ch
		}
	}

	return neg, prec, integer, fraction
}

// 2384249x10^-3  mantissa=2384249 exp=-3 -> float64:2384.249
func atof64exact(mantissa uint64, exp int, neg bool) (f float64, ok bool) {
	if mantissa>>float64info.mantbits != 0 {
		return
	}
	f = float64(mantissa)
	if neg {
		f = -f
	}
	switch {
	case exp == 0:
		// an integer.
		return f, true
	// Exact integers are <= 10^15.
	// Exact powers of ten are <= 10^22.
	case exp > 0 && exp <= 15+22: // int * 10^k
		// If exponent is big but number of digits is not,
		// can move a few zeros into the integer part.
		if exp > 22 {
			f *= float64pow10[exp-22]
			exp = 22
		}
		if f > 1e15 || f < -1e15 {
			// the exponent was really too large.
			return
		}
		return f * float64pow10[exp], true
	case exp < 0 && exp >= -22: // int / 10^k
		return f / float64pow10[-exp], true
	}
	return
}
