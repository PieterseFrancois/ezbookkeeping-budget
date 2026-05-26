package core

// BudgetEndDay represents the last day of each budget period.
// 0 means end of calendar month (default); valid non-zero values are 15, 25, 28.
type BudgetEndDay uint8

const (
	BUDGET_END_DAY_DEFAULT BudgetEndDay = 0   // end of calendar month
	BUDGET_END_DAY_INVALID BudgetEndDay = 255 // sentinel for "not set"
)

// ValidBudgetEndDays lists all allowed non-default values plus 0 (end-of-month).
var ValidBudgetEndDays = []BudgetEndDay{0, 15, 25, 28}

// IsValidBudgetEndDay returns true if d is one of the permitted values.
func IsValidBudgetEndDay(d BudgetEndDay) bool {
	for _, v := range ValidBudgetEndDays {
		if d == v {
			return true
		}
	}
	return false
}
