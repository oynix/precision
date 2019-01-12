package prec

func addDeprecated(f1, f2 float64) float64 {
	neg1, integer1, fraction1 := float642Bytes(f1)
	neg2, integer2, fraction2 := float642Bytes(f2)
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

	// integer
	var sumI = integer1
	var shorti = integer2
	lb1i := len(integer1)
	lb2i := len(integer2)
	lsi := lb2i

	if lb2i > lb1i {
		sumI = integer2
		shorti = integer1
		lsi = lb1i
	}
	for i := 0; i < lsi; i++ {
		sumI[i] += shorti[i]
	}

	// fraction
	var sumF = fraction1
	var shortf = fraction2
	lb1f := len(fraction1)
	lb2f := len(fraction2)
	lsf := lb2f

	if lb2f > lb1f {
		sumF = fraction2
		shortf = fraction1
		lsf = lb1f
	}
	for i := 0; i < lsf; i++ {
		sumF[i] += shortf[i]
	}

	// combine fraction
	llf := len(sumF)
	for i := llf - 1; i > 0; i-- {
		b := sumF[i]
		if b > 9 {
			sumF[i] = b % 10
			sumF[i-1] += b / 10
		}
	}
	tenth := sumF[0] // 首位小数
	if tenth > 9 {
		sumF[0] = tenth % 10
		sumI[0] += tenth / 10
	}

	// combine integer
	lli := len(sumI)
	for i := 0; i < lli-1; i++ {
		b := sumI[i]
		if b > 9 {
			sumI[i] = b % 10
			sumI[i+1] += b / 10
		}
	}
	mx := sumI[lli-1] // 整数最高位
	if mx > 9 {
		sumI[lli-1] = mx % 10
		sumI = append(sumI, mx/10)
	}

	f, _ := bytes2Float64(neg, sumI, sumF)
	return f
}

func readFloat(s string) (mantissa uint64, exp int, neg, trunc, ok bool) {
	const uint64digits = 19
	i := 0

	// optional sign
	if i >= len(s) {
		return
	}
	switch {
	case s[i] == '+':
		i++
	case s[i] == '-':
		neg = true
		i++
	}

	// digits
	sawdot := false
	sawdigits := false
	nd := 0
	ndMant := 0
	dp := 0
	for ; i < len(s); i++ {
		switch c := s[i]; true {
		case c == '.':
			if sawdot {
				return
			}
			sawdot = true
			dp = nd
			continue

		case '0' <= c && c <= '9':
			sawdigits = true
			if c == '0' && nd == 0 { // ignore leading zeros
				dp--
				continue
			}
			nd++
			if ndMant < uint64digits {
				mantissa *= 10
				mantissa += uint64(c - '0')
				ndMant++
			} else if s[i] != '0' {
				trunc = true
			}
			continue
		}
		break
	}
	if !sawdigits {
		return
	}
	if !sawdot {
		dp = nd
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 100000).  it doesn't matter if it's
	// not the exact number.
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i >= len(s) {
			return
		}
		esign := 1
		if s[i] == '+' {
			i++
		} else if s[i] == '-' {
			i++
			esign = -1
		}
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return
		}
		e := 0
		for ; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
			if e < 10000 {
				e = e*10 + int(s[i]) - '0'
			}
		}
		dp += e * esign
	}

	if i != len(s) {
		return
	}

	if mantissa != 0 {
		exp = dp - ndMant
	}
	ok = true
	return

}