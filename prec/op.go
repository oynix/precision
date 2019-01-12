package prec

func Add(f1, f2 float64) float64 {
	if f1 == 0 {
		return f2
	} else if f2 == 0 {
		return f1
	}
	neg1 := f1 < 0
	neg2 := f2 < 0
	neg := false
	if neg1 && neg2 { // all is negative
		neg = true
	} else if !neg1 && !neg2 { // all is not negative
		neg = false
	} else if neg1 && !neg2 { // 1 is negative, 2 is not
		return Sub(f2, -f1)
	} else if !neg1 && neg2 { // 1 is not negative, 2 is
		return Sub(f1, -f2)
	}
	_, prec1, integer1, fraction1 := float642Uints(f1)
	_, prec2, integer2, fraction2 := float642Uints(f2)
	// fraction
	max := max(prec1, prec2)
	sumF := fraction1*uint64pow10[max-prec1] + fraction2*uint64pow10[max-prec2]
	// integer
	sumI := integer1 + integer2 + sumF/uint64pow10[max]

	sumF = sumF % uint64pow10[max]

	sum := float64(sumI) + float64(sumF)/float64pow10[max]
	if neg {
		sum *= -1
	}
	return sum
}

func Sub(f1, f2 float64) float64 {
	if f1 == f2 {
		return 0
	} else if f1 == 0 {
		return -f2
	} else if f2 == 0 {
		return f1
	}
	neg1 := f1 < 0
	neg2 := f2 < 0
	if !neg1 && neg2 {
		// 2.3 - (-1.2) = 2.3 + (-(-1.2)) = 2.3 + 1.2
		return Add(f1, -f2)
	} else if neg1 && !neg2 {
		// -2.3 - 1.2 = -2.3 + (-1.2)
		return Add(f1, -f2)
	} else if neg1 && neg2 {
		// -2.3 - (-1.2) = -(-1.2) - 2.3 = 1.2 - 2.3
		return Sub(-f2, f1)
	}
	neg := f2 > f1
	var (
		prec1, prec2 int
		integer1, integer2, fraction1, fraction2 uint64
	)
	if neg {
		_, prec1, integer1, fraction1 = float642Uints(f2)
		_, prec2, integer2, fraction2 = float642Uints(f1)
	} else {
		_, prec1, integer1, fraction1 = float642Uints(f1)
		_, prec2, integer2, fraction2 = float642Uints(f2)
	}

	// fraction
	max := max(prec1, prec2)
	fraction1 = fraction1 * uint64pow10[max-prec1]
	fraction2 = fraction2 * uint64pow10[max-prec2]
	if fraction1 < fraction2 {
		fraction1 += uint64pow10[max]
		integer1 -= 1
	}
	diffF := fraction1 - fraction2

	// integer
	diffI := integer1 - integer2

	diff := float64(diffI) + float64(diffF)/float64pow10[max]
	if neg {
		diff *= -1
	}
	return diff
}
