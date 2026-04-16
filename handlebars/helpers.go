package handlebars

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FormatDate formats a date/datetime string as MM/DD/YYYY
func FormatDate(value any) string {
	if value == nil {
		return ""
	}

	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v == nil {
			return ""
		}
		t = *v
	case string:
		if v == "" {
			return ""
		}
		// Try RFC3339 first (ISO 8601)
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			// Try RFC3339Nano
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				// Try date only format
				t, err = time.Parse("2006-01-02", v)
				if err != nil {
					// Try datetime format with timezone
					t, err = time.Parse("2006-01-02T15:04:05Z07:00", v)
					if err != nil {
						return v // Return original if can't parse
					}
				}
			}
		}
	case int64:
		// Handle Unix timestamp
		// Check if it's in milliseconds (13+ digits) or seconds (10 digits)
		if v > 1000000000000 {
			// Milliseconds
			t = time.Unix(0, v*int64(time.Millisecond))
		} else {
			// Seconds
			t = time.Unix(v, 0)
		}
	case int32:
		// Handle Unix timestamp in seconds
		t = time.Unix(int64(v), 0)
	case int:
		// Handle Unix timestamp
		if v > 1000000000000 {
			// Milliseconds
			t = time.Unix(0, int64(v)*int64(time.Millisecond))
		} else {
			// Seconds
			t = time.Unix(int64(v), 0)
		}
	case float64:
		// Handle Unix timestamp
		if v > 1000000000000 {
			// Milliseconds
			t = time.Unix(0, int64(v)*int64(time.Millisecond))
		} else {
			// Seconds
			t = time.Unix(int64(v), 0)
		}
	default:
		// Try to convert to string and parse
		str := fmt.Sprintf("%v", value)

		// Check if it's a number string (epoch timestamp)
		if num, err := strconv.ParseInt(str, 10, 64); err == nil {
			if num > 1000000000000 {
				// Milliseconds
				t = time.Unix(0, num*int64(time.Millisecond))
			} else if num > 0 {
				// Seconds
				t = time.Unix(num, 0)
			} else {
				return str
			}
		} else {
			// Try parsing as RFC3339
			t, err = time.Parse(time.RFC3339, str)
			if err != nil {
				return str
			}
		}
	}

	// Convert to EST
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	t = t.In(loc)

	return t.Format("01/02/2006")
}

// FormatDateTime formats a date/datetime string as MM/DD/YYYY HH:MM AM/PM
func FormatDateTime(value any) string {
	if value == nil {
		return ""
	}

	var t time.Time
	var err error

	switch v := value.(type) {
	case time.Time:
		t = v
	case *time.Time:
		if v == nil {
			return ""
		}
		t = *v
	case string:
		if v == "" {
			return ""
		}
		// Try RFC3339 first (ISO 8601)
		t, err = time.Parse(time.RFC3339, v)
		if err != nil {
			// Try RFC3339Nano
			t, err = time.Parse(time.RFC3339Nano, v)
			if err != nil {
				// Try date only format
				t, err = time.Parse("2006-01-02", v)
				if err != nil {
					// Try datetime format with timezone
					t, err = time.Parse("2006-01-02T15:04:05Z07:00", v)
					if err != nil {
						return v // Return original if can't parse
					}
				}
			}
		}
	case int64:
		// Handle Unix timestamp
		// Check if it's in milliseconds (13+ digits) or seconds (10 digits)
		if v > 1000000000000 {
			// Milliseconds
			t = time.Unix(0, v*int64(time.Millisecond))
		} else {
			// Seconds
			t = time.Unix(v, 0)
		}
	case int32:
		// Handle Unix timestamp in seconds
		t = time.Unix(int64(v), 0)
	case int:
		// Handle Unix timestamp
		if v > 1000000000000 {
			// Milliseconds
			t = time.Unix(0, int64(v)*int64(time.Millisecond))
		} else {
			// Seconds
			t = time.Unix(int64(v), 0)
		}
	case float64:
		// Handle Unix timestamp
		if v > 1000000000000 {
			// Milliseconds
			t = time.Unix(0, int64(v)*int64(time.Millisecond))
		} else {
			// Seconds
			t = time.Unix(int64(v), 0)
		}
	default:
		// Try to convert to string and parse
		str := fmt.Sprintf("%v", value)

		// Check if it's a number string (epoch timestamp)
		if num, err := strconv.ParseInt(str, 10, 64); err == nil {
			if num > 1000000000000 {
				// Milliseconds
				t = time.Unix(0, num*int64(time.Millisecond))
			} else if num > 0 {
				// Seconds
				t = time.Unix(num, 0)
			} else {
				return str
			}
		} else {
			// Try parsing as RFC3339
			t, err = time.Parse(time.RFC3339, str)
			if err != nil {
				return str
			}
		}
	}

	// Convert to EST
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.UTC
	}
	t = t.In(loc)

	return t.Format("01/02/2006 03:04 PM")
}

// FormatCurrency formats a number as $X.XX
func FormatCurrency(value any) string {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case float32:
		num = float64(v)
	case int:
		num = float64(v)
	case int64:
		num = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return "$0.00"
		}
		num = parsed
	default:
		return "$0.00"
	}

	return fmt.Sprintf("$%.2f", num)
}

// Add performs addition
func Add(a, b any) float64 {
	return toFloat64(a) + toFloat64(b)
}

// Subtract performs subtraction
func Subtract(a, b any) float64 {
	return toFloat64(a) - toFloat64(b)
}

// Multiply performs multiplication
func Multiply(a, b any) float64 {
	return toFloat64(a) * toFloat64(b)
}

// Divide performs division (returns 0 if divisor is 0)
func Divide(a, b any) float64 {
	divisor := toFloat64(b)
	if divisor == 0 {
		return 0
	}
	return toFloat64(a) / divisor
}

// Percentage calculates (a / b) * 100
func Percentage(a, b any) float64 {
	divisor := toFloat64(b)
	if divisor == 0 {
		return 0
	}
	return (toFloat64(a) / divisor) * 100
}

// Round rounds a number to specified decimal places
func Round(value any, decimals any) float64 {
	val := toFloat64(value)
	dec := int(toFloat64(decimals))

	multiplier := math.Pow(10, float64(dec))
	return math.Round(val*multiplier) / multiplier
}

// Max returns the larger of two numbers
func Max(a, b any) float64 {
	aVal := toFloat64(a)
	bVal := toFloat64(b)
	if aVal > bVal {
		return aVal
	}
	return bVal
}

// Min returns the smaller of two numbers
func Min(a, b any) float64 {
	aVal := toFloat64(a)
	bVal := toFloat64(b)
	if aVal < bVal {
		return aVal
	}
	return bVal
}

// SecureSSN masks an SSN to show only last 4 digits
func SecureSSN(value any) string {
	if value == nil {
		return ""
	}

	ssn := fmt.Sprintf("%v", value)
	ssn = strings.TrimSpace(ssn)

	if ssn == "" {
		return ""
	}

	// Remove any dashes for processing
	cleanSSN := strings.ReplaceAll(ssn, "-", "")

	// If it's 9 digits, format as xxx-xx-1234
	if len(cleanSSN) == 9 {
		lastFour := cleanSSN[len(cleanSSN)-4:]
		return fmt.Sprintf("xxx-xx-%s", lastFour)
	}

	// If already formatted with dashes (###-##-####)
	if len(ssn) == 11 && ssn[3] == '-' && ssn[6] == '-' {
		lastFour := ssn[7:]
		return fmt.Sprintf("xxx-xx-%s", lastFour)
	}

	return ssn
}

// RunningBalance calculates remaining balance after each payment
// params: totalDebt (float64), payments ([]map with "amount" key), currentIndex (int)
// Returns float64 instead of string so it can be used with format_currency
func RunningBalance(totalDebt any, payments any, currentIndex any) float64 {
	debt := toFloat64(totalDebt)
	idx := int(toFloat64(currentIndex))

	// Handle payments array
	var paidSoFar float64

	// Use reflection to handle any slice type
	rv := reflect.ValueOf(payments)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		for i := 0; i < rv.Len(); i++ {
			if i <= idx {
				payment := rv.Index(i).Interface()

				// Try to extract amount from the payment
				var amount float64

				// Check if it's a map
				if paymentMap, ok := payment.(map[string]any); ok {
					// Try "amount" (lowercase - from JSON/BSON tags)
					if val, exists := paymentMap["amount"]; exists {
						amount = toFloat64(val)
					}
				} else if paymentMap, ok := payment.(map[string]any); ok {
					// Try "amount" (lowercase)
					if val, exists := paymentMap["amount"]; exists {
						amount = toFloat64(val)
					}
				} else {
					// Try to use reflection to get the Amount field
					paymentValue := reflect.ValueOf(payment)
					if paymentValue.Kind() == reflect.Struct {
						amountField := paymentValue.FieldByName("Amount")
						if amountField.IsValid() {
							amount = toFloat64(amountField.Interface())
						}
					} else if paymentValue.Kind() == reflect.Map {
						// Try map access with reflection
						amountKey := reflect.ValueOf("amount")
						amountValue := paymentValue.MapIndex(amountKey)
						if amountValue.IsValid() {
							amount = toFloat64(amountValue.Interface())
						}
					}
				}

				paidSoFar += amount
			}
		}
	}

	remaining := debt - paidSoFar
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// Count returns the length/count of an array, slice, string, or map
func Count(value any) int {
	if value == nil {
		return 0
	}

	// Use reflection to handle all slice, array, map, and string types
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
		return rv.Len()
	default:
		return 0
	}
}

// Gt returns true if a > b (greater than)
func Gt(a, b any) bool {
	return toFloat64(a) > toFloat64(b)
}

// Lt returns true if a < b (less than)
func Lt(a, b any) bool {
	return toFloat64(a) < toFloat64(b)
}

// Gte returns true if a >= b (greater than or equal)
func Gte(a, b any) bool {
	return toFloat64(a) >= toFloat64(b)
}

// Lte returns true if a <= b (less than or equal)
func Lte(a, b any) bool {
	return toFloat64(a) <= toFloat64(b)
}

// Eq returns true if a == b (equal)
func Eq(a, b any) bool {
	return toFloat64(a) == toFloat64(b)
}

// Neq returns true if a != b (not equal)
func Neq(a, b any) bool {
	return toFloat64(a) != toFloat64(b)
}

// toFloat64 converts various types to float64
func toFloat64(value any) float64 {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case int16:
		return float64(v)
	case int8:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case uint16:
		return float64(v)
	case uint8:
		return float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0
		}
		return parsed
	default:
		// Try to convert using fmt.Sprintf and parse
		str := fmt.Sprintf("%v", v)
		parsed, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return 0
		}
		return parsed
	}
}
