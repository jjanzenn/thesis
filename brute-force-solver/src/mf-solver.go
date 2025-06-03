package main

import (
	"fmt"
	"os"

	"git.jjanzen.ca/jjanzen/thesis/brute-force-solver/pkg/fraction"
	"git.jjanzen.ca/jjanzen/thesis/brute-force-solver/pkg/solver"
)

func readArgs() (uint8, []fraction.Fraction, error) {
	maxPrecision := 0
	errString := fmt.Errorf("")
	errOccured := false
	fractions := make([]fraction.Fraction, 0)
	for _, arg := range os.Args[1:] {
		frac, err := fraction.NewFraction(arg)
		if err != nil {
			errString = fmt.Errorf("%s%s\n", errString, err)
			errOccured = true
		} else {
			frac.Reduce()
			if maxPrecision < int(frac.DenominatorExponent) {
				maxPrecision = int(frac.DenominatorExponent)
			}
			fractions = append(fractions, frac)
		}
	}

	if errOccured {
		return 0, nil, fmt.Errorf("Failed to parse arguments: %s", errString)
	}

	maxPrecision += (len(fractions) - 1) / 4

	return uint8(maxPrecision), fractions, nil
}

func main() {
	maxPrecision, target, err := readArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, "Mixing targets ")
	for _, targ := range target {
		fmt.Fprintf(os.Stderr, "%s ", targ)
	}
	fmt.Fprintf(os.Stderr, "with maximum denominator %d\n", uint64(1<<uint64(maxPrecision)))

	list, err := solver.Solve(maxPrecision, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "No such graph: %s", err)
		os.Exit(1)
	}
	fmt.Fprintln(os.Stderr, list)
}
