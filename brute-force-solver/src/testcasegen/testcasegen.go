package main

import (
	"fmt"
	"os"
	"strconv"

	"git.jjanzen.ca/jjanzen/thesis/brute-force-solver/pkg/fraction"
)

func printCase(currCase []fraction.Fraction) {
	for _, frac := range currCase {
		newFrac := frac
		newFrac.Reduce()
		fmt.Printf("%s ", newFrac)
	}
	fmt.Println()
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "maximum graph width and maximum denominator exponent must be provided\n")
		os.Exit(1)
	}

	graphWidth, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil || graphWidth < 1 {
		fmt.Fprintf(os.Stderr, "cannot parse \"%s\" as a positive integer graph width", os.Args[1])
		os.Exit(1)
	}
	maxPrecision, err := strconv.ParseUint(os.Args[2], 10, 8)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse \"%s\" as a non-negative integer denominator exponent\n", os.Args[2])
		os.Exit(1)
	}
	if maxPrecision >= 64 {
		fmt.Fprintf(os.Stderr, "Denominators of size 2^64 and larger are not supported\n")
		os.Exit(1)
	}

	currCase := make([]fraction.Fraction, graphWidth)
	for i := range currCase {
		currCase[i] = fraction.Fraction{
			Numerator:           0,
			DenominatorExponent: uint8(maxPrecision),
		}
	}

	index := len(currCase) - 1
	for currCase[0].Numerator < 1<<maxPrecision {
		for index > 0 && currCase[index].Numerator > 1<<maxPrecision {
			index -= 1
			currCase[index].Numerator += 1
		}
		index += 1
		for index < len(currCase) {
			currCase[index].Numerator = currCase[index-1].Numerator
			index += 1
		}
		index = len(currCase) - 1
		var sum uint64 = 0
		for _, frac := range currCase {
			sum += frac.Numerator
		}

		if sum%uint64(1<<uint64(maxPrecision)) == 0 {
			printCase(currCase)
		}
		currCase[index].Numerator += 1
	}
}
