// Code generated by github.com/STRockefeller/linqable Do NOT EDIT.

package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
)

// Substitutions for the specified Material.
type Substitutions []Substitution

// Scan implements database/sql Scanner interface.
func (sub *Substitutions) Scan(src interface{}) error {
	return ScanJSON(src, sub)
}

// Value implements database/sql/driver Valuer interface.
func (sub Substitutions) Value() (driver.Value, error) {
	return json.Marshal(sub)
}

// NewSubstitutions returns a new Substitutions from the input Substitution.
func NewSubstitutions(sub []Substitution) Substitutions {
	return sub
}

// RepeatSubstitutions generates a sequence that contains one repeated value.
func RepeatSubstitutions(element Substitution, count int) Substitutions {
	sub := NewSubstitutions([]Substitution{})
	for i := 0; i < count; i++ {
		sub = sub.Append(element)
	}
	return sub
}

// Where filters a sequence of values based on a predicate.
func (sub Substitutions) Where(predicate func(Substitution) bool) Substitutions {
	res := []Substitution{}
	for _, elem := range sub {
		if predicate(elem) {
			res = append(res, elem)
		}
	}
	return res
}

// Reverse inverts the order of the elements in a sequence.
func (sub Substitutions) Reverse(predicate func(Substitution) bool) Substitutions {
	res := NewSubstitutions(make([]Substitution, len(sub)))
	for i, j := 0, len(sub)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = sub[j], sub[i]
	}
	return res
}

// Contains determines whether a sequence contains a specified element.
func (sub Substitutions) Contains(target Substitution) bool {
	for _, elem := range sub {
		if reflect.DeepEqual(elem, target) {
			return true
		}
	}
	return false
}

// Count returns  a number that represents how many elements in the specified sequence satisfy a condition.
func (sub Substitutions) Count(predicate func(Substitution) bool) int {
	var count int
	for _, elem := range sub {
		if predicate(elem) {
			count++
		}
	}
	return count
}

// Any determines whether any element of a sequence satisfies a condition.
func (sub Substitutions) Any(predicate func(Substitution) bool) bool {
	for _, elem := range sub {
		if predicate(elem) {
			return true
		}
	}
	return false
}

// All determines whether all elements of a sequence satisfy a condition.
func (sub Substitutions) All(predicate func(Substitution) bool) bool {
	for _, elem := range sub {
		if predicate(elem) {
			continue
		} else {
			return false
		}
	}
	return true
}

// Append appends a value to the end of the sequence.
func (sub Substitutions) Append(newItem Substitution) Substitutions {
	return append(sub, newItem)
}

// Prepend adds a value to the beginning of the sequence.
func (sub Substitutions) Prepend(newItem Substitution) Substitutions {
	return append([]Substitution{newItem}, sub.ToSlice()...)
}

// Distinct returns distinct elements from a sequence by usubng the default equality comparer to compare values.
func (sub Substitutions) Distinct() Substitutions {
	res := sub.Empty()
	for _, elem := range sub {
		if !res.Contains(elem) {
			res = res.Append(elem)
		}
	}
	return res
}

// ElementAt returns the element at a specified index in a sequence.
func (sub Substitutions) ElementAt(index int) Substitution {
	if index >= len(sub) {
		panic("linq: ElementAt() out of index")
	}
	return sub[index]
}

// ElementAtOrDefault returns the element at a specified index in a sequence or a default value if the index is out of range.
func (sub Substitutions) ElementAtOrDefault(index int) Substitution {
	var defaultValue Substitution
	if index >= len(sub) {
		return defaultValue
	}
	return sub[index]
}

// Empty returns empty Substitutions that has the specified type argument.
func (sub Substitutions) Empty() Substitutions {
	return NewSubstitutions([]Substitution{})
}

// First returns the first element in a sequence that satisfies a specified condition.
func (sub Substitutions) First(predicate func(Substitution) bool) Substitution {
	if len(sub) <= 0 {
		panic("linq: First() empty set")
	}
	for _, elem := range sub {
		if predicate(elem) {
			return elem
		}
	}
	panic("linq: First() no match element in the slice")
}

// FirstOrDefault returns the first element of a sequence, or a default value if the sequence contains no elements.
func (sub Substitutions) FirstOrDefault(predicate func(Substitution) bool) Substitution {
	var defaultValue Substitution
	if len(sub) <= 0 {
		return defaultValue
	}
	for _, elem := range sub {
		if predicate(elem) {
			return elem
		}
	}
	return defaultValue
}

// Last returns the last element in a sequence that satisfies a specified condition.
func (sub Substitutions) Last(predicate func(Substitution) bool) Substitution {
	if len(sub) <= 0 {
		panic("linq: Last() empty set")
	}
	for i := len(sub) - 1; i >= 0; i-- {
		if predicate(sub[i]) {
			return sub[i]
		}
	}
	panic("linq: Last() no match element in the slice")
}

// LastOrDefault returns the last element of a sequence, or a default value if the sequence contains no elements.
func (sub Substitutions) LastOrDefault(predicate func(Substitution) bool) Substitution {
	var defaultValue Substitution
	if len(sub) <= 0 {
		return defaultValue
	}
	for i := len(sub) - 1; i >= 0; i-- {
		if predicate(sub[i]) {
			return sub[i]
		}
	}
	return defaultValue
}

// Single returns the only element of a sequence that satisfies a specified condition, and returns a panic if more than one such element exists.
func (sub Substitutions) Single(predicate func(Substitution) bool) Substitution {
	if len(sub) <= 0 {
		panic("linq: Single() empty set")
	}
	if sub.Count(predicate) == 1 {
		return sub.First(predicate)
	}
	panic("linq: Single() eligible data count is not unique")
}

// SingleOrDefault returns the only element of a sequence, or a default value if the sequence is empty; this method returns a panic if there is more than one element in the sequence.
func (sub Substitutions) SingleOrDefault(predicate func(Substitution) bool) Substitution {
	var defaultValue Substitution
	if len(sub) <= 0 {
		return defaultValue
	}
	if sub.Count(predicate) == 1 {
		return sub.First(predicate)
	}
	panic("linq: SingleOrDefault() eligible data count is not unique")
}

// Take returns a specified number of contiguous elements from the start of a sequence.
func (sub Substitutions) Take(n int) Substitutions {
	if n < 0 || n >= len(sub) {
		panic("Linq: Take() out of index")
	}
	res := []Substitution{}
	for i := 0; i < n; i++ {
		res = append(res, sub[i])
	}
	return res
}

// TakeWhile returns elements from a sequence as long as a specified condition is true, and then skips the remaining elements.
func (sub Substitutions) TakeWhile(predicate func(Substitution) bool) Substitutions {
	res := []Substitution{}
	for i := 0; i < len(sub); i++ {
		if predicate(sub[i]) {
			res = append(res, sub[i])
		} else {
			return res
		}
	}
	return res
}

// TakeLast returns a new Substitutions that contains the last count elements from source.
func (sub Substitutions) TakeLast(n int) Substitutions {
	if n < 0 || n >= len(sub) {
		panic("Linq: TakeLast() out of index")
	}
	return sub.Skip(len(sub) - n)
}

// Skip bypasses a specified number of elements in a sequence and then returns the remaining elements.
func (sub Substitutions) Skip(n int) Substitutions {
	if n < 0 || n >= len(sub) {
		panic("Linq: Skip() out of index")
	}
	return sub[n:]
}

// SkipWhile bypasses elements in a sequence as long as a specified condition is true and then returns the remaining elements.
func (sub Substitutions) SkipWhile(predicate func(Substitution) bool) Substitutions {
	for i := 0; i < len(sub); i++ {
		if predicate(sub[i]) {
			continue
		} else {
			return sub[i:]
		}
	}
	return Substitutions{}
}

// SkipLast returns a new enumerable collection that contains the elements from source with the last count elements of the source collection omitted.
func (sub Substitutions) SkipLast(n int) Substitutions {
	if n < 0 || n > len(sub) {
		panic("Linq: SkipLast() out of index")
	}
	return sub.Take(len(sub) - n)
}

// SumInt32 computes the sum of a sequence of numeric values.
func (sub Substitutions) SumInt32(selector func(Substitution) int32) int32 {
	var sum int32
	for _, elem := range sub {
		sum += selector(elem)
	}
	return sum
}

// SumInt16 computes the sum of a sequence of numeric values.
func (sub Substitutions) SumInt16(selector func(Substitution) int16) int16 {
	var sum int16
	for _, elem := range sub {
		sum += selector(elem)
	}
	return sum
}

// SumInt64 computes the sum of a sequence of numeric values.
func (sub Substitutions) SumInt64(selector func(Substitution) int64) int64 {
	var sum int64
	for _, elem := range sub {
		sum += selector(elem)
	}
	return sum
}

// SumFloat32 computes the sum of a sequence of numeric values.
func (sub Substitutions) SumFloat32(selector func(Substitution) float32) float32 {
	var sum float32
	for _, elem := range sub {
		sum += selector(elem)
	}
	return sum
}

// SumFloat64 computes the sum of a sequence of numeric values.
func (sub Substitutions) SumFloat64(selector func(Substitution) float64) float64 {
	var sum float64
	for _, elem := range sub {
		sum += selector(elem)
	}
	return sum
}

// ToSlice creates a/an Substitutions from a/an Substitution.
func (sub Substitutions) ToSlice() []Substitution {
	return sub
}

// #region not linq

// ForEach executes the provided callback once for each element present in the Substitutions in ascending order.
func (sub Substitutions) ForEach(callBack func(Substitution)) {
	for _, elem := range sub {
		callBack(elem)
	}
}

// ReplaceAll replaces all oldValues with newValues in the Substitutions.
func (sub Substitutions) ReplaceAll(oldValue, newValue Substitution) Substitutions {
	res := NewSubstitutions([]Substitution{})
	for _, elem := range sub {
		if reflect.DeepEqual(elem, oldValue) {
			res = res.Append(newValue)
		} else {
			res = res.Append(elem)
		}
	}
	return res
}

// Remove removes the first occurrence of a specific object from the Substitutions.
func (sub *Substitutions) Remove(item Substitution) bool {
	res := NewSubstitutions([]Substitution{})
	var isRemoved bool
	for _, elem := range *sub {
		if reflect.DeepEqual(elem, item) && !isRemoved {
			isRemoved = true
			continue
		}
		res = res.Append(elem)
	}
	*sub = res
	return isRemoved
}

// RemoveAll removes all the elements that match the conditions defined by the specified predicate.
func (sub *Substitutions) RemoveAll(predicate func(Substitution) bool) int {
	var count int
	res := NewSubstitutions([]Substitution{})
	for _, elem := range *sub {
		if predicate(elem) {
			count++
			continue
		}
		res = res.Append(elem)
	}
	*sub = res
	return count
}

// RemoveAt removes the element at the specified index of the Substitutions.
func (sub *Substitutions) RemoveAt(index int) {
	res := NewSubstitutions([]Substitution{})
	for i := 0; i < len(*sub); i++ {
		if i == index {
			continue
		}
		res = res.Append((*sub)[i])
	}
	*sub = res
}

//RemoveRange removes a range of elements from the Substitutions.
func (sub *Substitutions) RemoveRange(index, count int) error {
	if index < 0 || count < 0 || index+count > len(*sub) {
		return fmt.Errorf("argument out of range")
	}
	res := NewSubstitutions([]Substitution{})
	for i := 0; i < len(*sub); i++ {
		if i >= index && count != 0 {
			count--
			continue
		}
		res = res.Append((*sub)[i])
	}
	*sub = res
	return nil
}

// #endregion not linq
