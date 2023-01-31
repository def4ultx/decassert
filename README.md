# Decimal Assert

Easier assert [shopspring/decimal](https://github.com/shopspring/decimal) in large struct without failed from different exponential.

## Usage

```go
import (
    "testing"

    "github.com/shopspring/decimal"
    "github.com/def4ultx/decassert"
)

func TestDecimal(t *testing.T) {
    type Data struct {
        value decimal.Decimal
    }

    exp := Data{decimal.NewFromFloat(97406.784)}
    act := Data{decimal.NewFromFloat(97406.784)}
    decassert.Equal(t, exp, act)
}
```

## How it work

`decassert.Equal` deep copy expected and actual value then set its decimal to zero and compare the result. If it pass
it will get all underlying decimal including unexported field and compare.
