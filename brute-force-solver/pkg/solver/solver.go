package solver

import (
	"fmt"
	"sort"
	"sync"

	"git.jjanzen.ca/jjanzen/thesis/brute-force-solver/pkg/fraction"
)

var seenStates sync.Map

func solveRecursively(result chan [][]fraction.Fraction, maxPrecision uint8, targets string, state []fraction.Fraction) {
	if targets == fmt.Sprint(state) {
		returnVal := make([][]fraction.Fraction, 0)
		returnVal = append(returnVal, state)
		result <- returnVal
		return
	}

	childResultChan := make(chan [][]fraction.Fraction)
	numChildren := 0
	for i, frac1 := range state {
		for j, frac2 := range state[i+1:] {
			if frac1 != frac2 {
				mix, err := frac1.Mix(frac2)
				if err == nil && mix.DenominatorExponent <= maxPrecision {
					stateCopy := make([]fraction.Fraction, len(state))
					copy(stateCopy, state)
					stateCopy[i] = mix
					stateCopy[i+1+j] = mix
					sort.Slice(stateCopy, func(i2, j2 int) bool {
						return stateCopy[i2].LessThan(stateCopy[j2])
					})

					strStateCopy := fmt.Sprint(stateCopy)
					_, ok := seenStates.LoadOrStore(strStateCopy, true)
					if !ok {
						numChildren++
						go solveRecursively(childResultChan, maxPrecision, targets, stateCopy)
					}
				}
			}
		}
	}

	results := make([][]fraction.Fraction, 1)
	results[0] = state
	numDone := 0
	noPathToTarget := true
	for numDone < numChildren {
		returnValLen := 0
		select {
		case returnVal := <-childResultChan:
			results = append(results, returnVal...)
			numDone++
			returnValLen = len(returnVal)
		}
		if returnValLen > 0 {
			noPathToTarget = false
			break
		}
	}
	if noPathToTarget {
		results = nil
	}

	result <- results
}

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

	numOnes := uint64(sum) / uint64(1<<uint64(maxPrecision))
	initial := make([]fraction.Fraction, 0)
	for range numOnes {
		initial = append(initial, fraction.Fraction{
			Numerator:           1,
			DenominatorExponent: 0,
		})
	}
	for range len(targets) - int(numOnes) {
		initial = append(initial, fraction.Fraction{
			Numerator:           0,
			DenominatorExponent: 0,
		})
	}
	sort.Slice(targets, func(i, j int) bool {
		return targets[i].LessThan(targets[j])
	})
	sort.Slice(initial, func(i, j int) bool {
		return initial[i].LessThan(initial[j])
	})

	results := make(chan [][]fraction.Fraction)
	go solveRecursively(results, maxPrecision, fmt.Sprint(targets), initial)
	list := <-results

	if len(list) == 0 {
		return nil, fmt.Errorf("no path to the target")
	}

	return list, nil
}
