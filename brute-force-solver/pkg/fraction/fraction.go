package fraction

import (
	"fmt"
	"strconv"
	"strings"
)

// stores a fraction of the form a/(2^x)
// breaks for x >= 64
type Fraction struct {
	Numerator           uint64
	DenominatorExponent uint8
}

func (f Fraction) String() string {
	if f.DenominatorExponent == 0 || f.Numerator == 0 {
		return fmt.Sprint(f.Numerator)
	}
	return fmt.Sprintf("%d/%d", f.Numerator, uint64(1<<f.DenominatorExponent))
}

func (f *Fraction) SetPrecision(p uint8) error {
	orig_p := f.DenominatorExponent
	f.Reduce()
	if p >= 64 {
		f.SetPrecision(orig_p)
		return fmt.Errorf("precision of %d is out of bounds", p)
	} else if p < f.DenominatorExponent {
		f.SetPrecision(orig_p)
		return fmt.Errorf("cannot decrease precision any further")
	} else {
		f.Numerator <<= (p - f.DenominatorExponent)
		f.DenominatorExponent = p
	}

	return nil
}

func (f1 Fraction) Mix(f2 Fraction) (Fraction, error) {
	var err error
	if f1.DenominatorExponent < f2.DenominatorExponent {
		err = f1.SetPrecision(f2.DenominatorExponent)
	} else if f1.DenominatorExponent > f2.DenominatorExponent {
		err = f2.SetPrecision(f1.DenominatorExponent)
	}

	if err != nil {
		return Fraction{0, 0}, fmt.Errorf("error mixing %s and %s: %e", f1, f2, err)
	}

	sum := f1.Numerator + f2.Numerator
	if sum < f1.Numerator || sum < f2.Numerator {
		return Fraction{0, 0}, fmt.Errorf("error mixing %s and %s: sum of numerators is out of bounds", f1, f2)
	}
	result := Fraction{
		Numerator:           sum,
		DenominatorExponent: f1.DenominatorExponent + 1,
	}
	result.Reduce()

	return result, nil
}

func (f *Fraction) Reduce() {
	if f.Numerator == 0 {
		f.DenominatorExponent = 0
	} else {
		for f.Numerator&1 == 0 && f.DenominatorExponent > 0 {
			f.DenominatorExponent -= 1
			f.Numerator >>= 1
		}
	}
}

func (f1 Fraction) LessThan(f2 Fraction) bool {
	if f1.DenominatorExponent > f2.DenominatorExponent {
		f2.SetPrecision(f1.DenominatorExponent)
	} else if f1.DenominatorExponent < f2.DenominatorExponent {
		f1.SetPrecision(f2.DenominatorExponent)
	}
	return f1.Numerator < f2.Numerator
}

func (f1 Fraction) Eq(f2 Fraction) bool {
	f1.Reduce()
	f2.Reduce()
	return f1 == f2
}

func NewFraction(str string) (Fraction, error) {
	parts := strings.Split(str, "/")
	if len(parts) > 2 || len(parts) == 0 {
		return Fraction{0, 0}, fmt.Errorf("cannot parse string \"%s\" as fraction", str)
	}

	var result Fraction

	numerator, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return Fraction{0, 0}, fmt.Errorf("cannot parse string \"%s\" as a fraction: numerator is not an integer", str)
	}
	result.Numerator = uint64(numerator)

	if len(parts) == 2 {
		denominator, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return Fraction{0, 0}, fmt.Errorf("cannot parse string \"%s\" as a fraction: denominator is not an integer", str)
		}
		if denominator == 0 {
			return Fraction{0, 0}, fmt.Errorf("division by zero in fraction \"%s\"", str)
		}
		denominatorExponent := 0
		for denominator&1 == 0 {
			denominatorExponent += 1
			denominator >>= 1
		}
		if denominator != 1 {
			return Fraction{0, 0}, fmt.Errorf("denominator of \"%s\" is not a power of two", str)
		}

		result.DenominatorExponent = uint8(denominatorExponent)
	} else {
		result.DenominatorExponent = 0
	}

	return result, nil
}
