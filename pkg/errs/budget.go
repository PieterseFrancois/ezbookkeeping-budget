package errs

import "net/http"

// Error codes related to budget targets
var (
	ErrBudgetTargetIdInvalid     = NewNormalError(NormalSubcategoryBudget, 0, http.StatusBadRequest, "budget target id is invalid")
	ErrBudgetTargetNotFound      = NewNormalError(NormalSubcategoryBudget, 1, http.StatusBadRequest, "budget target not found")
	ErrBudgetTargetAlreadyExists = NewNormalError(NormalSubcategoryBudget, 2, http.StatusBadRequest, "budget target already exists for this category and period")
)
