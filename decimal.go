package decassert

import (
	"fmt"
	"reflect"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type TestingT interface {
	Errorf(format string, args ...any)
}

func DecimalEqual(t TestingT, expected, actual decimal.Decimal, msgAndArgs ...any) bool {
	if !expected.Equal(actual) {
		assert.Fail(t, fmt.Sprintf("Expected %v; Actual %v; Diff %v}", expected, actual, expected.Sub(actual)), msgAndArgs...)
		return false
	}
	return true
}

func Equal(t TestingT, expected, actual any, msgAndArgs ...any) bool {
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, fmt.Sprintf("panic: %v", r))
		}
	}()

	skipCopyTypes := SkipCopyTypes(decimal.Decimal{})
	exp := DeepCopy(expected, skipCopyTypes)
	act := DeepCopy(actual, skipCopyTypes)
	if !assert.Equal(t, exp, act, msgAndArgs...) {
		return false
	}

	type Assertion struct {
		expected reflect.Value
		actual   reflect.Value
	}
	var (
		iterQueue = make([]Assertion, 0)
		visited   = make(map[uintptr]bool)

		ev = reflect.ValueOf(expected)
		av = reflect.ValueOf(actual)
	)

	iterQueue = append(iterQueue, Assertion{ev, av})

	for len(iterQueue) > 0 {
		pop := iterQueue[0]
		iterQueue = iterQueue[1:]

		if !pop.expected.IsValid() || !pop.actual.IsValid() {
			continue
		}
		if pop.expected.IsZero() && pop.actual.IsZero() {
			continue
		}

		if isDecimal(pop.expected, pop.actual) {
			var exp, act decimal.Decimal
			if pop.expected.CanInterface() && pop.actual.CanInterface() {
				exp = pop.expected.Interface().(decimal.Decimal)
				act = pop.actual.Interface().(decimal.Decimal)
			} else {
				exp = valueInterfaceUnsafe(pop.expected).(decimal.Decimal)
				act = valueInterfaceUnsafe(pop.actual).(decimal.Decimal)

			}
			ok := DecimalEqual(t, exp, act, msgAndArgs...)
			if !ok {
				return false
			}
			continue
		}

		switch pop.expected.Kind() {
		case reflect.Ptr:
			eAddr := uintptr(pop.expected.UnsafePointer())
			aAddr := uintptr(pop.actual.UnsafePointer())
			if marked, ok := visited[eAddr]; marked || ok {
				continue
			}
			if marked, ok := visited[aAddr]; marked || ok {
				continue
			}
			visited[eAddr] = true
			visited[aAddr] = true

			e := Assertion{
				actual:   pop.actual.Elem(),
				expected: pop.expected.Elem(),
			}
			iterQueue = append(iterQueue, e)
		case reflect.Struct:
			fields := pop.expected.Type()
			for i := 0; i < fields.NumField(); i++ {
				e := Assertion{
					actual:   pop.actual.Field(i),
					expected: pop.expected.Field(i),
				}
				iterQueue = append(iterQueue, e)
			}
		// No need to check map/slice size, already check by assert.Equal above
		case reflect.Slice, reflect.Array:
			for i := 0; i < pop.expected.Len(); i++ {
				e := Assertion{
					actual:   pop.actual.Index(i),
					expected: pop.expected.Index(i),
				}
				iterQueue = append(iterQueue, e)
			}
		case reflect.Map:
			for _, k := range pop.expected.MapKeys() {
				e := Assertion{
					actual:   pop.expected.MapIndex(k),
					expected: pop.actual.MapIndex(k),
				}
				iterQueue = append(iterQueue, e)
			}
		case reflect.Interface:
			e := Assertion{
				actual:   pop.expected.Elem(),
				expected: pop.actual.Elem(),
			}
			iterQueue = append(iterQueue, e)
		}
	}
	return true
}

var typeDecimal = reflect.TypeOf(decimal.Decimal{})

func isDecimal(a, b reflect.Value) bool {
	return a.Type() == typeDecimal && b.Type() == typeDecimal
}
