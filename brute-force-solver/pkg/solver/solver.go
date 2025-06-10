package solver

import (
	"fmt"
	"slices"
	"sort"
	"sync"

	"git.jjanzen.ca/jjanzen/thesis/brute-force-solver/pkg/fraction"
)

var seenStates sync.Map

type ErrorTree struct {
	state    []fraction.Fraction
	err      error
	children []ErrorTree
}

func (err *ErrorTree) errorTreeStringHelper(prelude string) string {
	result := fmt.Sprintf("%s%s %s\n", prelude, err.state, err.err)

	new_prelude := ""
	for _, c := range prelude {
		if c == '├' {
			c = '│'
		} else if c == '└' {
			c = ' '
		}
		new_prelude += string(c)
	}
	prelude = new_prelude

	if len(err.children) >= 2 {
		for _, child := range err.children[:len(err.children)-1] {
			result += child.errorTreeStringHelper(prelude + "├")
		}
	}
	if len(err.children) >= 1 {
		result += err.children[len(err.children)-1].errorTreeStringHelper(prelude + "└")
	}

	return result
}

func (err ErrorTree) String() string {
	return err.errorTreeStringHelper("")
}

func minCounts(state []fraction.Fraction, targetFracs []fraction.Fraction) (int, int) {
	if len(state) != len(targetFracs) || len(state) == 0 {
		return 0, 0
	}

	min := targetFracs[0]

	targetMinCount := 0
	for _, frac := range targetFracs {
		if frac.Eq(min) {
			targetMinCount++
		}
	}

	stateMinCount := 0
	for _, frac := range state {
		if frac.Eq(min) {
			stateMinCount++
		}
	}

	return stateMinCount, targetMinCount
}

func maxCounts(state []fraction.Fraction, targetFracs []fraction.Fraction) (int, int) {
	if len(state) != len(targetFracs) || len(state) == 0 {
		return 0, 0
	}

	max := targetFracs[len(targetFracs)-1]

	targetMaxCount := 0
	for _, frac := range targetFracs {
		if frac.Eq(max) {
			targetMaxCount++
		}
	}

	stateMaxCount := 0
	for _, frac := range state {
		if frac.Eq(max) {
			stateMaxCount++
		}
	}

	return stateMaxCount, targetMaxCount
}

func minMix(state []fraction.Fraction, goal fraction.Fraction) (fraction.Fraction, error) {
	// assumes state is sorted
	mix := fraction.Fraction{
		Numerator:           0,
		DenominatorExponent: 0,
	}

	// find the smallest value larger than the goal
	for _, frac := range state {
		if !frac.LessThan(goal) {
			mix = frac
			break
		}
	}

	// mix with every value less than the goal
	for _, frac := range state {
		if !goal.LessThan(frac) {
			newmix, err := mix.Mix(frac)
			if err != nil {
				return newmix, fmt.Errorf("cannot check correctness: %s\n", err)
			}
			mix = newmix
		}
	}

	return mix, nil
}

func maxMix(state []fraction.Fraction, goal fraction.Fraction) (fraction.Fraction, error) {
	// assumes state is sorted
	mix := fraction.Fraction{
		Numerator:           0,
		DenominatorExponent: 0,
	}

	// find the largest value smaller than the goal
	for _, frac := range slices.Backward(state) {
		if !goal.LessThan(frac) {
			mix = frac
			break
		}
	}

	// mix with every value greater than the goal
	for _, frac := range state {
		if !frac.LessThan(goal) {
			newmix, err := mix.Mix(frac)
			if err != nil {
				return newmix, fmt.Errorf("cannot check correctness: %s\n", err)
			}
			mix = newmix
		}
	}

	return mix, nil
}

func assertTargetIsReachable(state []fraction.Fraction, targetFracs []fraction.Fraction) error {
	// assumes target and state are sorted

	if len(state) != len(targetFracs) {
		// should never occur
		return fmt.Errorf("state does not have the same number of elements as the target")
	}
	if len(targetFracs) == 0 {
		return nil
	}

	stateMinCount, targetMinCount := minCounts(state, targetFracs)

	// insufficient number of the minimum fraction
	if !state[0].LessThan(targetFracs[0]) && stateMinCount < targetMinCount {
		return fmt.Errorf("insufficient instances of min value to reach target: %d < %d", stateMinCount, targetMinCount)
	}

	// no series of mixes will create the minimum
	mix, err := minMix(state, targetFracs[0])
	if err != nil {
		return err
	} else if stateMinCount < targetMinCount && targetFracs[0].LessThan(mix) {
		return fmt.Errorf("no mix will ever reach minimum: %s < %s", targetFracs[0], mix)
	}

	stateMaxCount, targetMaxCount := maxCounts(state, targetFracs)

	// insufficient number of the maximum fraction
	if !targetFracs[len(targetFracs)-1].LessThan(state[len(state)-1]) && stateMaxCount < targetMaxCount {
		return fmt.Errorf("insufficient instances of max value to reach target: %d < %d", stateMaxCount, targetMaxCount)
	}

	// no series of mixes will create the maximum
	mix, err = maxMix(state, targetFracs[len(targetFracs)-1])
	if err != nil {
		return err
	} else if stateMaxCount < targetMaxCount && mix.LessThan(targetFracs[len(targetFracs)-1]) {
		return fmt.Errorf("no mix will ever reach maximum: %s < %s", mix, targetFracs[len(targetFracs)-1])
	}

	return nil
}

func solveRecursively(
	result chan [][]fraction.Fraction,
	errors chan ErrorTree,
	maxPrecision uint8,
	targetFracs []fraction.Fraction,
	state []fraction.Fraction,
	beforeSaved []fraction.Fraction,
	afterSaved []fraction.Fraction,
) {
	if fmt.Sprint(targetFracs) == fmt.Sprint(state) {
		returnVal := append(beforeSaved, state...)
		returnVal = append(returnVal, afterSaved...)
		result <- [][]fraction.Fraction{returnVal}
		return
	}

	childResults := make(chan [][]fraction.Fraction)
	childErrors := make(chan ErrorTree)
	childErrorTree := make([]ErrorTree, 0)

	// create copy of the state with saved values before and after
	printableState := make([]fraction.Fraction, len(state))
	copy(printableState, state)
	printableState = append(beforeSaved, printableState...)
	printableState = append(printableState, afterSaved...)

	// count the number of child goroutines
	numChildren := 0

	for i, frac1 := range state {
		for j, frac2 := range state[i+1:] {
			if frac1 == frac2 {
				continue // don't do nops
			}

			// try mixing the two fractions, continuing loop early on error
			mix, err := frac1.Mix(frac2)
			if err != nil {
				childErrorTree = append(childErrorTree, ErrorTree{
					state:    nil,
					err:      err,
					children: nil,
				})
				continue
			}

			// perform the mixing operation on a copy of the state;
			// keep it sorted
			stateCopy := make([]fraction.Fraction, len(state))
			copy(stateCopy, state)
			stateCopy[i] = mix
			stateCopy[i+1+j] = mix
			sort.Slice(stateCopy, func(i2, j2 int) bool {
				return stateCopy[i2].LessThan(stateCopy[j2])
			})

			// check if the state has been seen already
			// otherwise, note it as seen
			_, ok := seenStates.LoadOrStore(fmt.Sprint(stateCopy), true)

			// create copies of the target and saved value lists
			targetFracsCopy := make([]fraction.Fraction, len(targetFracs))
			copy(targetFracsCopy, targetFracs)
			beforeSavedCopy := make([]fraction.Fraction, len(beforeSaved))
			copy(beforeSavedCopy, beforeSaved)
			afterSavedCopy := make([]fraction.Fraction, len(afterSaved))
			copy(afterSavedCopy, afterSaved)

			// trim values from the state that cannot be used
			// add them to the saved value lists
			for len(stateCopy) > 0 && stateCopy[0].Eq(targetFracsCopy[0]) {
				beforeSavedCopy = append(beforeSavedCopy, stateCopy[0])
				stateCopy = stateCopy[1:]
				targetFracsCopy = targetFracsCopy[1:]
			}
			for len(stateCopy) > 0 && stateCopy[len(stateCopy)-1].Eq(targetFracsCopy[len(targetFracsCopy)-1]) {
				afterSavedCopy = append([]fraction.Fraction{stateCopy[len(stateCopy)-1]}, afterSavedCopy...)
				stateCopy = stateCopy[:len(stateCopy)-1]
				targetFracsCopy = targetFracsCopy[:len(targetFracsCopy)-1]
			}

			// make a printable version of the state including the saved values
			newPrintableState := append(beforeSavedCopy, stateCopy...)
			newPrintableState = append(newPrintableState, afterSavedCopy...)

			// check for invariants
			err = assertTargetIsReachable(stateCopy, targetFracsCopy)

			// if a state has aleady been seen, an invariant is invalidated,
			// or the denominator is too large, append to the error tree
			if ok {
				childErrorTree = append(childErrorTree, ErrorTree{
					state:    newPrintableState,
					err:      fmt.Errorf("state already seen"),
					children: nil,
				})
			} else if err != nil {
				childErrorTree = append(childErrorTree, ErrorTree{
					state:    newPrintableState,
					err:      err,
					children: nil,
				})
			} else if mix.DenominatorExponent > maxPrecision {
				childErrorTree = append(childErrorTree, ErrorTree{
					state:    newPrintableState,
					err:      fmt.Errorf("denominator %d too large", 1<<mix.DenominatorExponent),
					children: nil,
				})
			} else {
				if !ok {
					numChildren++
					go solveRecursively(
						childResults,
						childErrors,
						maxPrecision,
						targetFracsCopy,
						stateCopy,
						beforeSavedCopy,
						afterSavedCopy,
					)
				}
			}
		}
	}

	results := make([][]fraction.Fraction, 1)
	results[0] = printableState

	numDone := 0
	noPathToTarget := true
	for numDone < numChildren {
		returnValLen := 0
		select {
		case returnVal := <-childResults:
			results = append(results, returnVal...)
			numDone++
			returnValLen = len(returnVal)
		case returnVal := <-childErrors:
			childErrorTree = append(childErrorTree, returnVal)
			numDone++
		}
		if returnValLen > 0 {
			noPathToTarget = false
			break
		}
	}

	if noPathToTarget {
		err := ErrorTree{
			state:    printableState,
			err:      fmt.Errorf(""),
			children: childErrorTree,
		}
		errors <- err
	} else {
		result <- results
	}
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
	errors := make(chan ErrorTree)
	go solveRecursively(results, errors, maxPrecision, targets, initial, nil, nil)

	list := make([][]fraction.Fraction, 0)
	var err ErrorTree
	select {
	case list = <-results:
		return list, nil
	case err = <-errors:
		return nil, fmt.Errorf("no path to the target:\n%s", err)
	}
}
