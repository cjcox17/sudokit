package handlebars

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatDate(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		result := FormatDate(nil)
		assert.Equal(t, "", result)
	})

	t.Run("TimeValue", func(t *testing.T) {
		date := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
		result := FormatDate(date)
		assert.Contains(t, result, "/15/2024")
	})

	t.Run("TimePointer", func(t *testing.T) {
		date := time.Date(2024, 3, 25, 14, 45, 0, 0, time.UTC)
		result := FormatDate(&date)
		assert.Contains(t, result, "/25/2024")
	})

	t.Run("NilTimePointer", func(t *testing.T) {
		var date *time.Time = nil
		result := FormatDate(date)
		assert.Equal(t, "", result)
	})

	t.Run("RFC3339String", func(t *testing.T) {
		result := FormatDate("2024-01-15T10:30:00Z")
		assert.Contains(t, result, "/15/2024")
	})

	t.Run("DateOnlyString", func(t *testing.T) {
		result := FormatDate("2024-12-25")
		// Date conversion may vary by timezone, just check it's formatted
		assert.Contains(t, result, "/2024")
		assert.Len(t, result, 10) // MM/DD/YYYY format
	})

	t.Run("EmptyString", func(t *testing.T) {
		result := FormatDate("")
		assert.Equal(t, "", result)
	})

	t.Run("InvalidString", func(t *testing.T) {
		result := FormatDate("not a date")
		assert.Equal(t, "not a date", result)
	})

	t.Run("OtherType", func(t *testing.T) {
		result := FormatDate(12345)
		assert.Equal(t, "12/31/1969", result)
	})
}

func TestFormatDateTime(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		result := FormatDateTime(nil)
		assert.Equal(t, "", result)
	})

	t.Run("TimeValue", func(t *testing.T) {
		datetime := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
		result := FormatDateTime(datetime)
		assert.Contains(t, result, "/2024")
		// Time will vary based on EST conversion, just check format
		assert.Regexp(t, `\d{2}/\d{2}/\d{4} \d{2}:\d{2} (AM|PM)`, result)
	})

	t.Run("TimePointer", func(t *testing.T) {
		datetime := time.Date(2024, 3, 25, 9, 15, 0, 0, time.UTC)
		result := FormatDateTime(&datetime)
		assert.Contains(t, result, "/25/2024")
		assert.Contains(t, result, "AM")
	})

	t.Run("NilTimePointer", func(t *testing.T) {
		var datetime *time.Time = nil
		result := FormatDateTime(datetime)
		assert.Equal(t, "", result)
	})

	t.Run("RFC3339String", func(t *testing.T) {
		result := FormatDateTime("2024-01-15T14:30:00Z")
		assert.Contains(t, result, "/15/2024")
	})

	t.Run("DateOnlyString", func(t *testing.T) {
		result := FormatDateTime("2024-12-25")
		// Date conversion may vary by timezone
		assert.Contains(t, result, "/2024")
		assert.Regexp(t, `\d{2}/\d{2}/\d{4} \d{2}:\d{2} (AM|PM)`, result)
	})

	t.Run("EmptyString", func(t *testing.T) {
		result := FormatDateTime("")
		assert.Equal(t, "", result)
	})

	t.Run("InvalidString", func(t *testing.T) {
		result := FormatDateTime("invalid datetime")
		assert.Equal(t, "invalid datetime", result)
	})

	t.Run("OtherType", func(t *testing.T) {
		result := FormatDateTime(99999)
		assert.Equal(t, "01/01/1970 10:46 PM", result)
	})
}

func TestFormatCurrency(t *testing.T) {
	t.Run("Float64", func(t *testing.T) {
		result := FormatCurrency(123.45)
		assert.Equal(t, "$123.45", result)
	})

	t.Run("Float32", func(t *testing.T) {
		result := FormatCurrency(float32(99.99))
		assert.Equal(t, "$99.99", result)
	})

	t.Run("Integer", func(t *testing.T) {
		result := FormatCurrency(100)
		assert.Equal(t, "$100.00", result)
	})

	t.Run("Int64", func(t *testing.T) {
		result := FormatCurrency(int64(250))
		assert.Equal(t, "$250.00", result)
	})

	t.Run("StringNumber", func(t *testing.T) {
		result := FormatCurrency("75.50")
		assert.Equal(t, "$75.50", result)
	})

	t.Run("InvalidString", func(t *testing.T) {
		result := FormatCurrency("not a number")
		assert.Equal(t, "$0.00", result)
	})

	t.Run("ZeroValue", func(t *testing.T) {
		result := FormatCurrency(0)
		assert.Equal(t, "$0.00", result)
	})

	t.Run("NegativeValue", func(t *testing.T) {
		result := FormatCurrency(-50.25)
		assert.Equal(t, "$-50.25", result)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		result := FormatCurrency([]int{1, 2, 3})
		assert.Equal(t, "$0.00", result)
	})

	t.Run("LargePrecision", func(t *testing.T) {
		result := FormatCurrency(123.456789)
		assert.Equal(t, "$123.46", result)
	})
}

func TestAdd(t *testing.T) {
	t.Run("TwoIntegers", func(t *testing.T) {
		result := Add(5, 3)
		assert.Equal(t, 8.0, result)
	})

	t.Run("TwoFloats", func(t *testing.T) {
		result := Add(10.5, 2.3)
		assert.InDelta(t, 12.8, result, 0.01)
	})

	t.Run("MixedTypes", func(t *testing.T) {
		result := Add(5, 2.5)
		assert.Equal(t, 7.5, result)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Add("10", 5)
		assert.Equal(t, 15.0, result)
	})

	t.Run("WithNil", func(t *testing.T) {
		result := Add(nil, 5)
		assert.Equal(t, 5.0, result)
	})

	t.Run("BothNil", func(t *testing.T) {
		result := Add(nil, nil)
		assert.Equal(t, 0.0, result)
	})
}

func TestSubtract(t *testing.T) {
	t.Run("TwoIntegers", func(t *testing.T) {
		result := Subtract(10, 3)
		assert.Equal(t, 7.0, result)
	})

	t.Run("TwoFloats", func(t *testing.T) {
		result := Subtract(10.5, 2.3)
		assert.InDelta(t, 8.2, result, 0.01)
	})

	t.Run("NegativeResult", func(t *testing.T) {
		result := Subtract(5, 10)
		assert.Equal(t, -5.0, result)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Subtract("20", 5)
		assert.Equal(t, 15.0, result)
	})

	t.Run("WithNil", func(t *testing.T) {
		result := Subtract(10, nil)
		assert.Equal(t, 10.0, result)
	})
}

func TestMultiply(t *testing.T) {
	t.Run("TwoIntegers", func(t *testing.T) {
		result := Multiply(5, 3)
		assert.Equal(t, 15.0, result)
	})

	t.Run("TwoFloats", func(t *testing.T) {
		result := Multiply(2.5, 4.0)
		assert.Equal(t, 10.0, result)
	})

	t.Run("WithZero", func(t *testing.T) {
		result := Multiply(100, 0)
		assert.Equal(t, 0.0, result)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Multiply("3", 7)
		assert.Equal(t, 21.0, result)
	})

	t.Run("WithNil", func(t *testing.T) {
		result := Multiply(5, nil)
		assert.Equal(t, 0.0, result)
	})
}

func TestDivide(t *testing.T) {
	t.Run("TwoIntegers", func(t *testing.T) {
		result := Divide(10, 2)
		assert.Equal(t, 5.0, result)
	})

	t.Run("TwoFloats", func(t *testing.T) {
		result := Divide(10.0, 4.0)
		assert.Equal(t, 2.5, result)
	})

	t.Run("DivideByZero", func(t *testing.T) {
		result := Divide(10, 0)
		assert.Equal(t, 0.0, result)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Divide("20", 4)
		assert.Equal(t, 5.0, result)
	})

	t.Run("WithNilDivisor", func(t *testing.T) {
		result := Divide(10, nil)
		assert.Equal(t, 0.0, result)
	})

	t.Run("NilDividend", func(t *testing.T) {
		result := Divide(nil, 5)
		assert.Equal(t, 0.0, result)
	})
}

func TestPercentage(t *testing.T) {
	t.Run("ValidPercentage", func(t *testing.T) {
		result := Percentage(25, 100)
		assert.Equal(t, 25.0, result)
	})

	t.Run("FiftyPercent", func(t *testing.T) {
		result := Percentage(50, 100)
		assert.Equal(t, 50.0, result)
	})

	t.Run("OverHundred", func(t *testing.T) {
		result := Percentage(150, 100)
		assert.Equal(t, 150.0, result)
	})

	t.Run("DivideByZero", func(t *testing.T) {
		result := Percentage(10, 0)
		assert.Equal(t, 0.0, result)
	})

	t.Run("WithFloats", func(t *testing.T) {
		result := Percentage(33.33, 100.0)
		assert.InDelta(t, 33.33, result, 0.01)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Percentage("20", "80")
		assert.Equal(t, 25.0, result)
	})
}

func TestRound(t *testing.T) {
	t.Run("RoundToTwoDecimals", func(t *testing.T) {
		result := Round(123.456, 2)
		assert.Equal(t, 123.46, result)
	})

	t.Run("RoundToZeroDecimals", func(t *testing.T) {
		result := Round(123.456, 0)
		assert.Equal(t, 123.0, result)
	})

	t.Run("RoundToOneDecimal", func(t *testing.T) {
		result := Round(123.456, 1)
		assert.Equal(t, 123.5, result)
	})

	t.Run("RoundNegativeNumber", func(t *testing.T) {
		result := Round(-123.456, 2)
		assert.Equal(t, -123.46, result)
	})

	t.Run("RoundWithString", func(t *testing.T) {
		result := Round("99.999", 2)
		assert.Equal(t, 100.0, result)
	})

	t.Run("NoRoundingNeeded", func(t *testing.T) {
		result := Round(100.0, 2)
		assert.Equal(t, 100.0, result)
	})
}

func TestMax(t *testing.T) {
	t.Run("FirstIsLarger", func(t *testing.T) {
		result := Max(10, 5)
		assert.Equal(t, 10.0, result)
	})

	t.Run("SecondIsLarger", func(t *testing.T) {
		result := Max(5, 10)
		assert.Equal(t, 10.0, result)
	})

	t.Run("BothEqual", func(t *testing.T) {
		result := Max(7, 7)
		assert.Equal(t, 7.0, result)
	})

	t.Run("WithNegatives", func(t *testing.T) {
		result := Max(-5, -10)
		assert.Equal(t, -5.0, result)
	})

	t.Run("WithFloats", func(t *testing.T) {
		result := Max(3.14, 3.15)
		assert.Equal(t, 3.15, result)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Max("50", 25)
		assert.Equal(t, 50.0, result)
	})
}

func TestMin(t *testing.T) {
	t.Run("FirstIsSmaller", func(t *testing.T) {
		result := Min(5, 10)
		assert.Equal(t, 5.0, result)
	})

	t.Run("SecondIsSmaller", func(t *testing.T) {
		result := Min(10, 5)
		assert.Equal(t, 5.0, result)
	})

	t.Run("BothEqual", func(t *testing.T) {
		result := Min(7, 7)
		assert.Equal(t, 7.0, result)
	})

	t.Run("WithNegatives", func(t *testing.T) {
		result := Min(-5, -10)
		assert.Equal(t, -10.0, result)
	})

	t.Run("WithFloats", func(t *testing.T) {
		result := Min(3.14, 3.15)
		assert.Equal(t, 3.14, result)
	})

	t.Run("WithString", func(t *testing.T) {
		result := Min("25", 50)
		assert.Equal(t, 25.0, result)
	})
}

func TestSecureSSN(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		result := SecureSSN(nil)
		assert.Equal(t, "", result)
	})

	t.Run("NineDigits", func(t *testing.T) {
		result := SecureSSN("123456789")
		assert.Equal(t, "xxx-xx-6789", result)
	})

	t.Run("FormattedSSN", func(t *testing.T) {
		result := SecureSSN("123-45-6789")
		assert.Equal(t, "xxx-xx-6789", result)
	})

	t.Run("EmptyString", func(t *testing.T) {
		result := SecureSSN("")
		assert.Equal(t, "", result)
	})

	t.Run("WithWhitespace", func(t *testing.T) {
		result := SecureSSN("  123456789  ")
		assert.Equal(t, "xxx-xx-6789", result)
	})

	t.Run("InvalidFormat", func(t *testing.T) {
		result := SecureSSN("12345")
		assert.Equal(t, "12345", result)
	})

	t.Run("IntegerValue", func(t *testing.T) {
		result := SecureSSN(123456789)
		assert.Equal(t, "xxx-xx-6789", result)
	})

	t.Run("NonStandardFormat", func(t *testing.T) {
		result := SecureSSN("123 45 6789")
		assert.Equal(t, "123 45 6789", result)
	})
}

func TestRunningBalance(t *testing.T) {
	t.Run("NoPayments", func(t *testing.T) {
		payments := []any{}
		result := RunningBalance(1000.0, payments, 0)
		assert.Equal(t, 1000.0, result)
	})

	t.Run("FirstPayment", func(t *testing.T) {
		payments := []any{
			map[string]any{"amount": 100.0},
			map[string]any{"amount": 150.0},
		}
		result := RunningBalance(1000.0, payments, 0)
		assert.Equal(t, 900.0, result)
	})

	t.Run("SecondPayment", func(t *testing.T) {
		payments := []any{
			map[string]any{"amount": 100.0},
			map[string]any{"amount": 150.0},
		}
		result := RunningBalance(1000.0, payments, 1)
		assert.Equal(t, 750.0, result)
	})

	t.Run("AllPaymentsMade", func(t *testing.T) {
		payments := []any{
			map[string]any{"amount": 300.0},
			map[string]any{"amount": 400.0},
			map[string]any{"amount": 300.0},
		}
		result := RunningBalance(1000.0, payments, 2)
		assert.Equal(t, 0.0, result)
	})

	t.Run("OverpaymentHandling", func(t *testing.T) {
		payments := []any{
			map[string]any{"amount": 600.0},
			map[string]any{"amount": 500.0},
		}
		result := RunningBalance(1000.0, payments, 1)
		assert.Equal(t, 0.0, result)
	})

	t.Run("MapSlicePayments", func(t *testing.T) {
		payments := []map[string]any{
			{"amount": 200.0},
			{"amount": 300.0},
		}
		result := RunningBalance(1000.0, payments, 0)
		assert.Equal(t, 800.0, result)
	})

	t.Run("WithStringAmount", func(t *testing.T) {
		payments := []any{
			map[string]any{"amount": "100.50"},
		}
		result := RunningBalance(500.0, payments, 0)
		assert.Equal(t, 399.5, result)
	})

	t.Run("WithMissingAmountField", func(t *testing.T) {
		payments := []any{
			map[string]any{"other": 100.0},
		}
		result := RunningBalance(500.0, payments, 0)
		assert.Equal(t, 500.0, result)
	})

	t.Run("NegativeIndex", func(t *testing.T) {
		payments := []any{
			map[string]any{"amount": 100.0},
		}
		result := RunningBalance(500.0, payments, -1)
		assert.Equal(t, 500.0, result)
	})
}

func TestToFloat64(t *testing.T) {
	t.Run("Float64", func(t *testing.T) {
		result := toFloat64(123.45)
		assert.Equal(t, 123.45, result)
	})

	t.Run("Float32", func(t *testing.T) {
		result := toFloat64(float32(99.99))
		assert.InDelta(t, 99.99, result, 0.01)
	})

	t.Run("Int", func(t *testing.T) {
		result := toFloat64(42)
		assert.Equal(t, 42.0, result)
	})

	t.Run("Int64", func(t *testing.T) {
		result := toFloat64(int64(100))
		assert.Equal(t, 100.0, result)
	})

	t.Run("Int32", func(t *testing.T) {
		result := toFloat64(int32(50))
		assert.Equal(t, 50.0, result)
	})

	t.Run("Uint", func(t *testing.T) {
		result := toFloat64(uint(75))
		assert.Equal(t, 75.0, result)
	})

	t.Run("Uint64", func(t *testing.T) {
		result := toFloat64(uint64(200))
		assert.Equal(t, 200.0, result)
	})

	t.Run("Uint32", func(t *testing.T) {
		result := toFloat64(uint32(150))
		assert.Equal(t, 150.0, result)
	})

	t.Run("ValidString", func(t *testing.T) {
		result := toFloat64("123.45")
		assert.Equal(t, 123.45, result)
	})

	t.Run("InvalidString", func(t *testing.T) {
		result := toFloat64("not a number")
		assert.Equal(t, 0.0, result)
	})

	t.Run("NilValue", func(t *testing.T) {
		result := toFloat64(nil)
		assert.Equal(t, 0.0, result)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		result := toFloat64([]int{1, 2, 3})
		assert.Equal(t, 0.0, result)
	})
}

func TestCount(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		result := Count(nil)
		assert.Equal(t, 0, result)
	})

	t.Run("Slice", func(t *testing.T) {
		result := Count([]int{1, 2, 3, 4, 5})
		assert.Equal(t, 5, result)
	})

	t.Run("EmptySlice", func(t *testing.T) {
		result := Count([]string{})
		assert.Equal(t, 0, result)
	})

	t.Run("StringSlice", func(t *testing.T) {
		result := Count([]string{"a", "b", "c"})
		assert.Equal(t, 3, result)
	})

	t.Run("InterfaceSlice", func(t *testing.T) {
		result := Count([]any{1, "two", 3.0})
		assert.Equal(t, 3, result)
	})

	t.Run("Array", func(t *testing.T) {
		arr := [4]int{1, 2, 3, 4}
		result := Count(arr)
		assert.Equal(t, 4, result)
	})

	t.Run("String", func(t *testing.T) {
		result := Count("hello")
		assert.Equal(t, 5, result)
	})

	t.Run("EmptyString", func(t *testing.T) {
		result := Count("")
		assert.Equal(t, 0, result)
	})

	t.Run("Map", func(t *testing.T) {
		m := map[string]int{"a": 1, "b": 2, "c": 3}
		result := Count(m)
		assert.Equal(t, 3, result)
	})

	t.Run("EmptyMap", func(t *testing.T) {
		result := Count(map[string]string{})
		assert.Equal(t, 0, result)
	})

	t.Run("NonCountableType", func(t *testing.T) {
		result := Count(42)
		assert.Equal(t, 0, result)
	})

	t.Run("NonCountableStruct", func(t *testing.T) {
		type testStruct struct {
			Field string
		}
		result := Count(testStruct{Field: "test"})
		assert.Equal(t, 0, result)
	})
}

func TestGt(t *testing.T) {
	t.Run("GreaterThan", func(t *testing.T) {
		assert.True(t, Gt(10, 5))
		assert.True(t, Gt(5.5, 5.4))
	})

	t.Run("NotGreaterThan", func(t *testing.T) {
		assert.False(t, Gt(5, 10))
		assert.False(t, Gt(5, 5))
	})

	t.Run("NilValues", func(t *testing.T) {
		assert.False(t, Gt(nil, nil))
		assert.True(t, Gt(1, nil))
		assert.False(t, Gt(nil, 1))
	})
}

func TestLt(t *testing.T) {
	t.Run("LessThan", func(t *testing.T) {
		assert.True(t, Lt(5, 10))
		assert.True(t, Lt(5.4, 5.5))
	})

	t.Run("NotLessThan", func(t *testing.T) {
		assert.False(t, Lt(10, 5))
		assert.False(t, Lt(5, 5))
	})

	t.Run("NilValues", func(t *testing.T) {
		assert.False(t, Lt(nil, nil))
		assert.False(t, Lt(1, nil))
		assert.True(t, Lt(nil, 1))
	})
}

func TestGte(t *testing.T) {
	t.Run("GreaterThanOrEqual", func(t *testing.T) {
		assert.True(t, Gte(10, 5))
		assert.True(t, Gte(5, 5))
		assert.True(t, Gte(5.5, 5.4))
	})

	t.Run("NotGreaterThanOrEqual", func(t *testing.T) {
		assert.False(t, Gte(5, 10))
	})
}

func TestLte(t *testing.T) {
	t.Run("LessThanOrEqual", func(t *testing.T) {
		assert.True(t, Lte(5, 10))
		assert.True(t, Lte(5, 5))
		assert.True(t, Lte(5.4, 5.5))
	})

	t.Run("NotLessThanOrEqual", func(t *testing.T) {
		assert.False(t, Lte(10, 5))
	})
}

func TestEq(t *testing.T) {
	t.Run("Equal", func(t *testing.T) {
		assert.True(t, Eq(5, 5))
		assert.True(t, Eq(5.5, 5.5))
		assert.True(t, Eq(0, 0))
	})

	t.Run("NotEqual", func(t *testing.T) {
		assert.False(t, Eq(5, 10))
		assert.False(t, Eq(5.5, 5.4))
	})

	t.Run("NilValues", func(t *testing.T) {
		assert.True(t, Eq(nil, nil))
		assert.False(t, Eq(1, nil))
		assert.False(t, Eq(nil, 1))
	})
}

func TestNeq(t *testing.T) {
	t.Run("NotEqual", func(t *testing.T) {
		assert.True(t, Neq(5, 10))
		assert.True(t, Neq(5.5, 5.4))
	})

	t.Run("Equal", func(t *testing.T) {
		assert.False(t, Neq(5, 5))
		assert.False(t, Neq(5.5, 5.5))
	})

	t.Run("NilValues", func(t *testing.T) {
		assert.False(t, Neq(nil, nil))
		assert.True(t, Neq(1, nil))
		assert.True(t, Neq(nil, 1))
	})
}
