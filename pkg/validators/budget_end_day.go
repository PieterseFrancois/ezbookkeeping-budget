package validators

import (
	"github.com/go-playground/validator/v10"

	"github.com/mayswind/ezbookkeeping/pkg/core"
)

// ValidateBudgetEndDay validates that the budget end day is one of the permitted values.
func ValidateBudgetEndDay(fl validator.FieldLevel) bool {
	d, ok := fl.Field().Interface().(core.BudgetEndDay)
	return ok && core.IsValidBudgetEndDay(d)
}
