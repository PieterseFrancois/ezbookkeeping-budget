package errs

import "net/http"

// Error codes related to goals
var (
	ErrGoalIdInvalid = NewNormalError(NormalSubcategoryGoal, 0, http.StatusBadRequest, "goal id is invalid")
	ErrGoalNotFound  = NewNormalError(NormalSubcategoryGoal, 1, http.StatusBadRequest, "goal not found")
)
