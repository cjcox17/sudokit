package handlebars

import (
	"fmt"

	"github.com/aymerick/raymond"
)

// RegisterHelpers registers all custom Handlebars helpers with raymond
func RegisterHelpers() {
	raymond.RegisterHelper("format_date", func(value any) raymond.SafeString {
		return raymond.SafeString(FormatDate(value))
	})

	raymond.RegisterHelper("format_datetime", func(value any) raymond.SafeString {
		return raymond.SafeString(FormatDateTime(value))
	})

	raymond.RegisterHelper("format_currency", func(value any) raymond.SafeString {
		return raymond.SafeString(FormatCurrency(value))
	})

	raymond.RegisterHelper("add", func(a, b any) string {
		return fmt.Sprintf("%v", Add(a, b))
	})

	raymond.RegisterHelper("subtract", func(a, b any) string {
		return fmt.Sprintf("%v", Subtract(a, b))
	})

	raymond.RegisterHelper("multiply", func(a, b any) string {
		return fmt.Sprintf("%v", Multiply(a, b))
	})

	raymond.RegisterHelper("divide", func(a, b any) string {
		return fmt.Sprintf("%v", Divide(a, b))
	})

	raymond.RegisterHelper("percentage", func(a, b any) string {
		return fmt.Sprintf("%.2f", Percentage(a, b))
	})

	raymond.RegisterHelper("round", func(value, decimals any) string {
		return fmt.Sprintf("%v", Round(value, decimals))
	})

	raymond.RegisterHelper("max", func(a, b any) string {
		return fmt.Sprintf("%v", Max(a, b))
	})

	raymond.RegisterHelper("min", func(a, b any) string {
		return fmt.Sprintf("%v", Min(a, b))
	})

	raymond.RegisterHelper("secure_ssn", func(value any) raymond.SafeString {
		return raymond.SafeString(SecureSSN(value))
	})

	raymond.RegisterHelper("running_balance", func(totalDebt, payments, currentIndex any) float64 {
		return RunningBalance(totalDebt, payments, currentIndex)
	})

	raymond.RegisterHelper("count", func(value any) string {
		return fmt.Sprintf("%d", Count(value))
	})

	raymond.RegisterHelper("gt", func(a, b any) bool {
		return Gt(a, b)
	})

	raymond.RegisterHelper("lt", func(a, b any) bool {
		return Lt(a, b)
	})

	raymond.RegisterHelper("gte", func(a, b any) bool {
		return Gte(a, b)
	})

	raymond.RegisterHelper("lte", func(a, b any) bool {
		return Lte(a, b)
	})

	raymond.RegisterHelper("eq", func(a, b any) bool {
		return Eq(a, b)
	})

	raymond.RegisterHelper("neq", func(a, b any) bool {
		return Neq(a, b)
	})
}
