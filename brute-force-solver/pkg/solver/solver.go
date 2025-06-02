package solver

import (
	"fmt"

	"git.jjanzen.ca/jjanzen/thesis/brute-force-solver/pkg/fraction"
)

func Solve(maxPrecision uint8, targets []fraction.Fraction) ([][]fraction.Fraction, error) {
	var sum uint64 = 0
	for _, target := range targets {
		target.SetPrecision(maxPrecision)
		sum += target.Numerator
		target.Reduce()
	}

	// TODO: this will be flaky for very large inputs
	if sum%uint64(1<<uint64(maxPrecision)) != 0 {
		return nil, fmt.Errorf("inputs do not sum to an integer")
	}

	return nil, nil
}
