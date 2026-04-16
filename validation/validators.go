package validation

import (
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

var Validator = validator.New()

func init() {
	// Register custom validations
	Validator.RegisterValidation("alpha_only", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		for _, r := range str {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == ' ' || r == '-' || r == '\'') {
				return false
			}
		}
		return true
	})

	Validator.RegisterValidation("alphanumeric_with_whitespace", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		for _, r := range str {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_') {
				return false
			}
		}
		return true
	})

	Validator.RegisterValidation("street_address", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		for _, r := range str {
			if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' || r == '/' || r == ',' || r == '#' || r == '.' || r == '(' || r == ')' || r == '&') {
				return false
			}
		}
		return true
	})

	Validator.RegisterValidation("ssn_format", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		if str == "" {
			return true
		}
		cleaned := strings.ReplaceAll(str, "-", "")
		if len(cleaned) != 9 {
			return false
		}
		for _, r := range cleaned {
			if !((r >= '0' && r <= '9') || r == 'X' || r == 'x') {
				return false
			}
		}
		return true
	})

	Validator.RegisterValidation("digits_only", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		for _, r := range str {
			if r < '0' || r > '9' {
				return false
			}
		}
		return true
	})

	Validator.RegisterValidation("us_zip_code", func(fl validator.FieldLevel) bool {
		str := fl.Field().String()
		// Match 5 digits or 5+4 format
		if len(str) == 5 {
			for _, r := range str {
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}
		if len(str) == 10 && str[5] == '-' {
			for i, r := range str {
				if i == 5 {
					continue
				}
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}
		return false
	})

	luhnCheckFunc := func(cardNumber string) bool {
		sum := 0
		isEven := false

		for i := len(cardNumber) - 1; i >= 0; i-- {
			n := int(cardNumber[i] - '0')
			if n < 0 || n > 9 {
				return false
			}

			if isEven {
				n = n * 2
				if n > 9 {
					n = n - 9
				}
			}

			sum += n
			isEven = !isEven
		}

		return sum%10 == 0
	}

	Validator.RegisterValidation("credit_card", func(fl validator.FieldLevel) bool {
		cardNumber := fl.Field().String()
		if len(cardNumber) < 13 || len(cardNumber) > 19 {
			return false
		}
		for _, r := range cardNumber {
			if r < '0' || r > '9' {
				return false
			}
		}
		return luhnCheckFunc(cardNumber)
	})

	Validator.RegisterValidation("future_date", func(fl validator.FieldLevel) bool {
		field := fl.Field()

		// Handle both *time.Time and time.Time
		var date *time.Time
		switch v := field.Interface().(type) {
		case *time.Time:
			date = v
		case time.Time:
			date = &v
		default:
			return false
		}

		if date == nil {
			return true
		}
		return date.After(time.Now())
	})
}
