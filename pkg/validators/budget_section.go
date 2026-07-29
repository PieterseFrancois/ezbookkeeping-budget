package validators

import (
	"github.com/go-playground/validator/v10"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ValidateBudgetSection validates that the budget section is one of the permitted values.
func ValidateBudgetSection(fl validator.FieldLevel) bool {
	s, ok := fl.Field().Interface().(string)
	return ok && models.IsValidBudgetSection(s)
}
