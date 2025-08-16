package array

import (
	"agg_go/internal/basics"
)

// IntLess compares two integers for less-than relationship.
// This is equivalent to AGG's int_less function.
func IntLess(a, b int) bool {
	return a < b
}

// IntGreater compares two integers for greater-than relationship.
// This is equivalent to AGG's int_greater function.
func IntGreater(a, b int) bool {
	return a > b
}

// UnsignedLess compares two unsigned integers for less-than relationship.
// This is equivalent to AGG's unsigned_less function.
func UnsignedLess(a, b uint) bool {
	return a < b
}

// UnsignedGreater compares two unsigned integers for greater-than relationship.
// This is equivalent to AGG's unsigned_greater function.
func UnsignedGreater(a, b uint) bool {
	return a > b
}

// Generic comparison functions for basic types

// Int8Less compares two int8 values.
func Int8Less(a, b basics.Int8) bool {
	return a < b
}

// Int8uLess compares two uint8 values.
func Int8uLess(a, b basics.Int8u) bool {
	return a < b
}

// Int16Less compares two int16 values.
func Int16Less(a, b basics.Int16) bool {
	return a < b
}

// Int16uLess compares two uint16 values.
func Int16uLess(a, b basics.Int16u) bool {
	return a < b
}

// Int32Less compares two int32 values.
func Int32Less(a, b basics.Int32) bool {
	return a < b
}

// Int32uLess compares two uint32 values.
func Int32uLess(a, b basics.Int32u) bool {
	return a < b
}

// Int64Less compares two int64 values.
func Int64Less(a, b basics.Int64) bool {
	return a < b
}

// Int64uLess compares two uint64 values.
func Int64uLess(a, b basics.Int64u) bool {
	return a < b
}

// Float32Less compares two float32 values.
func Float32Less(a, b float32) bool {
	return a < b
}

// Float64Less compares two float64 values.
func Float64Less(a, b float64) bool {
	return a < b
}

// Generic equality functions for basic types

// IntEqual compares two integers for equality.
func IntEqual(a, b int) bool {
	return a == b
}

// UnsignedEqual compares two unsigned integers for equality.
func UnsignedEqual(a, b uint) bool {
	return a == b
}

// Int8Equal compares two int8 values for equality.
func Int8Equal(a, b basics.Int8) bool {
	return a == b
}

// Int8uEqual compares two uint8 values for equality.
func Int8uEqual(a, b basics.Int8u) bool {
	return a == b
}

// Int16Equal compares two int16 values for equality.
func Int16Equal(a, b basics.Int16) bool {
	return a == b
}

// Int16uEqual compares two uint16 values for equality.
func Int16uEqual(a, b basics.Int16u) bool {
	return a == b
}

// Int32Equal compares two int32 values for equality.
func Int32Equal(a, b basics.Int32) bool {
	return a == b
}

// Int32uEqual compares two uint32 values for equality.
func Int32uEqual(a, b basics.Int32u) bool {
	return a == b
}

// Int64Equal compares two int64 values for equality.
func Int64Equal(a, b basics.Int64) bool {
	return a == b
}

// Int64uEqual compares two uint64 values for equality.
func Int64uEqual(a, b basics.Int64u) bool {
	return a == b
}

// Float32Equal compares two float32 values for equality.
func Float32Equal(a, b float32) bool {
	return a == b
}

// Float64Equal compares two float64 values for equality.
func Float64Equal(a, b float64) bool {
	return a == b
}

// Float32EqualEps compares two float32 values for equality within epsilon.
func Float32EqualEps(a, b, epsilon float32) bool {
	if a < b {
		return (b - a) <= epsilon
	}
	return (a - b) <= epsilon
}

// Float64EqualEps compares two float64 values for equality within epsilon.
func Float64EqualEps(a, b, epsilon float64) bool {
	return basics.IsEqualEps(a, b, epsilon)
}

// Comparator types for use with algorithms

// Comparator is a generic interface for comparison functions.
type Comparator[T any] interface {
	Less(a, b T) bool
	Equal(a, b T) bool
}

// IntComparator provides comparison functions for integers.
type IntComparator struct{}

// Less compares two integers.
func (IntComparator) Less(a, b int) bool {
	return a < b
}

// Equal compares two integers for equality.
func (IntComparator) Equal(a, b int) bool {
	return a == b
}

// UintComparator provides comparison functions for unsigned integers.
type UintComparator struct{}

// Less compares two unsigned integers.
func (UintComparator) Less(a, b uint) bool {
	return a < b
}

// Equal compares two unsigned integers for equality.
func (UintComparator) Equal(a, b uint) bool {
	return a == b
}

// Float64Comparator provides comparison functions for float64 values.
type Float64Comparator struct {
	Epsilon float64
}

// NewFloat64Comparator creates a new float64 comparator with the specified epsilon.
func NewFloat64Comparator(epsilon float64) Float64Comparator {
	return Float64Comparator{Epsilon: epsilon}
}

// Less compares two float64 values.
func (fc Float64Comparator) Less(a, b float64) bool {
	return a < b
}

// Equal compares two float64 values for equality within epsilon.
func (fc Float64Comparator) Equal(a, b float64) bool {
	return Float64EqualEps(a, b, fc.Epsilon)
}

// StringComparator provides comparison functions for strings.
type StringComparator struct{}

// Less compares two strings lexicographically.
func (StringComparator) Less(a, b string) bool {
	return a < b
}

// Equal compares two strings for equality.
func (StringComparator) Equal(a, b string) bool {
	return a == b
}

// ReverseComparator wraps another comparator to reverse the comparison order.
type ReverseComparator[T any] struct {
	base Comparator[T]
}

// NewReverseComparator creates a new reverse comparator.
func NewReverseComparator[T any](base Comparator[T]) ReverseComparator[T] {
	return ReverseComparator[T]{base: base}
}

// Less compares two values in reverse order.
func (rc ReverseComparator[T]) Less(a, b T) bool {
	return rc.base.Less(b, a)
}

// Equal compares two values for equality (unchanged).
func (rc ReverseComparator[T]) Equal(a, b T) bool {
	return rc.base.Equal(a, b)
}

// Helper functions to create comparison functions from comparators

// LessFuncFromComparator creates a LessFunc from a Comparator.
func LessFuncFromComparator[T any](comp Comparator[T]) LessFunc[T] {
	return comp.Less
}

// EqualFuncFromComparator creates an EqualFunc from a Comparator.
func EqualFuncFromComparator[T any](comp Comparator[T]) EqualFunc[T] {
	return comp.Equal
}

// Global comparator instances for convenience
var (
	DefaultIntComparator     = IntComparator{}
	DefaultUintComparator    = UintComparator{}
	DefaultFloat64Comparator = NewFloat64Comparator(1e-10)
	DefaultStringComparator  = StringComparator{}
)

// Utility functions for common sorting scenarios

// SortInts sorts a slice of integers in ascending order.
func SortInts(slice []int) {
	QuickSortSlice(slice, IntLess)
}

// SortIntsDescending sorts a slice of integers in descending order.
func SortIntsDescending(slice []int) {
	QuickSortSlice(slice, IntGreater)
}

// SortUints sorts a slice of unsigned integers in ascending order.
func SortUints(slice []uint) {
	QuickSortSlice(slice, UnsignedLess)
}

// SortUintsDescending sorts a slice of unsigned integers in descending order.
func SortUintsDescending(slice []uint) {
	QuickSortSlice(slice, UnsignedGreater)
}

// SortFloat64s sorts a slice of float64 values in ascending order.
func SortFloat64s(slice []float64) {
	QuickSortSlice(slice, Float64Less)
}

// SortStrings sorts a slice of strings in lexicographic order.
func SortStrings(slice []string) {
	QuickSortSlice(slice, func(a, b string) bool { return a < b })
}

// IsSortedInts checks if a slice of integers is sorted in ascending order.
func IsSortedInts(slice []int) bool {
	return IsSortedSlice(slice, IntLess)
}

// IsSortedUints checks if a slice of unsigned integers is sorted in ascending order.
func IsSortedUints(slice []uint) bool {
	return IsSortedSlice(slice, UnsignedLess)
}

// IsSortedFloat64s checks if a slice of float64 values is sorted in ascending order.
func IsSortedFloat64s(slice []float64) bool {
	return IsSortedSlice(slice, Float64Less)
}

// IsSortedStrings checks if a slice of strings is sorted in lexicographic order.
func IsSortedStrings(slice []string) bool {
	return IsSortedSlice(slice, func(a, b string) bool { return a < b })
}
