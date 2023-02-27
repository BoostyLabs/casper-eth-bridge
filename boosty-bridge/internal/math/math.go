package math

import (
	"errors"
)

// Range consists of 2 numbers and describes the range between them.
type Range struct {
	From uint64
	To   uint64
}

// SplitRange splits a large range into smaller ones.
func SplitRange(r Range, splitByNumber uint64) ([]Range, error) {
	var (
		splittedRange   []Range
		isBreakFromLoop bool
	)

	if r.From > r.To {
		return nil, errors.New("failed, to less than from")
	}

	if splitByNumber == 0 {
		return nil, errors.New("failed, split by number equal to 0")
	}

	for {
		internalTo := r.To
		amount := r.To - r.From

		switch {
		case amount <= 0:
			isBreakFromLoop = true
		case amount > splitByNumber:
			internalTo = r.From + splitByNumber

			splittedRange = append(splittedRange, Range{r.From, internalTo})
			r.From += splitByNumber
		case amount <= splitByNumber:
			splittedRange = append(splittedRange, Range{r.From, internalTo})

			return splittedRange, nil
		}

		if isBreakFromLoop {
			break
		}
	}

	return splittedRange, nil
}
