package prec

type extFloat struct {
	mant uint64
	exp  int
	neg  bool
}

func (f *extFloat) AssignComputeBounds(mant uint64, exp int, neg bool, flt *floatInfo) (lower, upper extFloat) {
	f.mant = mant
	f.exp = exp - int(flt.mantbits)
	f.neg = neg
	if f.exp <= 0 && mant == (mant>>uint(-f.exp))<<uint(-f.exp) {
		// An exact integer
		f.mant >>= uint(-f.exp)
		f.exp = 0
		return *f, *f
	}
	expBiased := exp - flt.bias

	upper = extFloat{mant: 2*f.mant + 1, exp: f.exp - 1, neg: f.neg}
	if mant != 1<<flt.mantbits || expBiased == 1 {
		lower = extFloat{mant: 2*f.mant - 1, exp: f.exp - 1, neg: f.neg}
	} else {
		lower = extFloat{mant: 4*f.mant - 1, exp: f.exp - 2, neg: f.neg}
	}
	return
}

func (f *extFloat) ShortestDecimal(d *decimalSlice, lower, upper *extFloat) bool {
	if f.mant == 0 {
		d.nd = 0
		d.dp = 0
		d.neg = f.neg
		return true
	}

	if f.exp == 0 && *lower == *f && *lower == *upper {
		// an exact integer.
		var buf [24]byte
		n := len(buf) - 1
		for v := f.mant; v > 0; {
			v1 := v / 10
			v -= 10 * v1
			buf[n] = byte(v + '0')
			n--
			v = v1
		}
		nd := len(buf) - n - 1
		for i := 0; i < nd; i++ {
			d.d[i] = buf[n+1+i]
		}
		d.nd, d.dp = nd, nd
		for d.nd > 0 && d.d[d.nd-1] == '0' {
			d.nd--
		}
		if d.nd == 0 {
			d.dp = 0
		}
		d.neg = f.neg
		return true
	}

	upper.Normalize()
	// Uniformize exponents.
	if f.exp > upper.exp {
		f.mant <<= uint(f.exp - upper.exp)
		f.exp = upper.exp
	}
	if lower.exp > upper.exp {
		lower.mant <<= uint(lower.exp - upper.exp)
		lower.exp = upper.exp
	}

	exp10 := frexp10Many(lower, f, upper)
	// Take a safety margin due to rounding in frexp10Many, but we lose precision.
	upper.mant++
	lower.mant--

	// The shortest representation of f is either rounded up or down, but
	// in any case, it is a truncation of upper.
	shift := uint(-upper.exp)
	integer := uint32(upper.mant >> shift)
	fraction := upper.mant - (uint64(integer) << shift)

	// How far we can go down from upper until the result is wrong.
	allowance := upper.mant - lower.mant
	// How far we should go to get a very precise result.
	targetDiff := upper.mant - f.mant

	// Count integral digits: there are at most 10.
	var integerDigits int
	for i, pow := 0, uint64(1); i < 20; i++ {
		if pow > uint64(integer) {
			integerDigits = i
			break
		}
		pow *= 10
	}
	for i := 0; i < integerDigits; i++ {
		pow := uint64pow10[integerDigits-i-1]

		digit := integer / uint32(pow)
		d.d[i] = byte(digit + '0')
		integer -= digit * uint32(pow)
		// evaluate whether we should stop.
		if currentDiff := uint64(integer)<<shift + fraction; currentDiff < allowance {
			d.nd = i + 1
			d.dp = integerDigits + exp10
			d.neg = f.neg
			// Sometimes allowance is so large the last digit might need to be
			// decremented to get closer to f.
			return adjustLastDigit(d, currentDiff, targetDiff, allowance, pow<<shift, 2)
		}
	}
	d.nd = integerDigits
	d.dp = d.nd + exp10
	d.neg = f.neg

	// Compute digits of the fractional part. At each step fraction does not
	// overflow. The choice of minExp implies that fraction is less than 2^60.
	var digit int
	multiplier := uint64(1)
	for {
		fraction *= 10
		multiplier *= 10
		digit = int(fraction >> shift)
		d.d[d.nd] = byte(digit + '0')
		d.nd++
		fraction -= uint64(digit) << shift
		if fraction < allowance*multiplier {
			// We are in the admissible range. Note that if allowance is about to
			// overflow, that is, allowance > 2^64/10, the condition is automatically
			// true due to the limited range of fraction.
			return adjustLastDigit(d,
				fraction, targetDiff*multiplier, allowance*multiplier,
				1<<shift, multiplier*2)
		}
	}
}

func (f *extFloat) Normalize() (shift uint) {
	mant, exp := f.mant, f.exp
	if mant == 0 {
		return 0
	}
	if mant>>(64-32) == 0 {
		mant <<= 32
		exp -= 32
	}
	if mant>>(64-16) == 0 {
		mant <<= 16
		exp -= 16
	}
	if mant>>(64-8) == 0 {
		mant <<= 8
		exp -= 8
	}
	if mant>>(64-4) == 0 {
		mant <<= 4
		exp -= 4
	}
	if mant>>(64-2) == 0 {
		mant <<= 2
		exp -= 2
	}
	if mant>>(64-1) == 0 {
		mant <<= 1
		exp -= 1
	}
	shift = uint(f.exp - exp)
	f.mant, f.exp = mant, exp
	return
}

func (f *extFloat) Multiply(g extFloat) {
	fhi, flo := f.mant>>32, uint64(uint32(f.mant))
	ghi, glo := g.mant>>32, uint64(uint32(g.mant))

	// Cross products.
	cross1 := fhi * glo
	cross2 := flo * ghi

	// f.mant*g.mant is fhi*ghi << 64 + (cross1+cross2) << 32 + flo*glo
	f.mant = fhi*ghi + (cross1 >> 32) + (cross2 >> 32)
	rem := uint64(uint32(cross1)) + uint64(uint32(cross2)) + ((flo * glo) >> 32)
	// Round up.
	rem += (1 << 31)

	f.mant += (rem >> 32)
	f.exp = f.exp + g.exp + 64
}

func (f *extFloat) frexp10() (exp10, index int) {
	// The constants expMin and expMax constrain the final value of the
	// binary exponent of f. We want a small integral part in the result
	// because finding digits of an integer requires divisions, whereas
	// digits of the fractional part can be found by repeatedly multiplying
	// by 10.
	const expMin = -60
	const expMax = -32
	// Find power of ten such that x * 10^n has a binary exponent
	// between expMin and expMax.
	approxExp10 := ((expMin+expMax)/2 - f.exp) * 28 / 93 // log(10)/log(2) is close to 93/28.
	i := (approxExp10 - firstPowerOfTen) / stepPowerOfTen
Loop:
	for {
		exp := f.exp + powersOfTen[i].exp + 64
		switch {
		case exp < expMin:
			i++
		case exp > expMax:
			i--
		default:
			break Loop
		}
	}
	// Apply the desired decimal shift on f. It will have exponent
	// in the desired range. This is multiplication by 10^-exp10.
	f.Multiply(powersOfTen[i])

	return -(firstPowerOfTen + i*stepPowerOfTen), i
}
