package models

// BudgetTarget represents budget target data stored in database.
// Section makes a target direction-specific, so one category can hold independent targets
// (e.g. a savings category budgeting both a contribution and an expected withdrawal).
type BudgetTarget struct {
	Id         int64  `xorm:"PK"`
	Uid        int64  `xorm:"UNIQUE(IDX_budget_uid_category_section_year_month) NOT NULL"`
	CategoryId int64  `xorm:"UNIQUE(IDX_budget_uid_category_section_year_month) NOT NULL"`
	Section    string `xorm:"VARCHAR(10) UNIQUE(IDX_budget_uid_category_section_year_month) NOT NULL DEFAULT ''"`
	Year       int    `xorm:"UNIQUE(IDX_budget_uid_category_section_year_month) NOT NULL"`
	Month      int    `xorm:"UNIQUE(IDX_budget_uid_category_section_year_month) NOT NULL"`
	Amount     int64  `xorm:"NOT NULL"`
}

// BudgetTargetCreateRequest represents all parameters of budget target creation request
type BudgetTargetCreateRequest struct {
	CategoryId int64  `json:"categoryId,string" binding:"required,min=1"`
	Section    string `json:"section" binding:"omitempty,validBudgetSection"`
	Year       int    `json:"year" binding:"required,min=1"`
	Month      int    `json:"month" binding:"required,min=1,max=12"`
	Amount     int64  `json:"amount,string" binding:"min=0"`
}

// BudgetTargetUpdateRequest represents all parameters of budget target modification request
type BudgetTargetUpdateRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Amount int64 `json:"amount,string" binding:"min=0"`
}

// BudgetTargetInfoResponse represents a view-object of budget target
type BudgetTargetInfoResponse struct {
	Id         int64  `json:"id,string"`
	CategoryId int64  `json:"categoryId,string"`
	Section    string `json:"section"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Amount     int64  `json:"amount,string"`
}

// BudgetTargetsGetRequest represents all parameters of budget targets listing request
type BudgetTargetsGetRequest struct {
	Year       int   `form:"year" binding:"required,min=1"`
	Month      int   `form:"month" binding:"required,min=1,max=12"`
	CategoryId int64 `form:"category_id,string" binding:"min=0"`
}

// BudgetTargetDeleteRequest represents all parameters of budget target deleting request
type BudgetTargetDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// Budget section names for the unified budget actuals. Each maps 1:1 to a section on the budget page.
const (
	BUDGET_SECTION_INCOME  = "income"
	BUDGET_SECTION_EXPENSE = "expense"
	BUDGET_SECTION_SAVINGS = "savings"
	BUDGET_SECTION_DEBT    = "debt"
)

// IsValidBudgetSection reports whether the given value is a known budget section name
func IsValidBudgetSection(section string) bool {
	switch section {
	case BUDGET_SECTION_INCOME, BUDGET_SECTION_EXPENSE, BUDGET_SECTION_SAVINGS, BUDGET_SECTION_DEBT:
		return true
	default:
		return false
	}
}

// Transfer parent category names that determine which section a transfer target belongs to.
// These mirror the constants used by the budget pages in the frontend.
const (
	SavingsTransferParentName = "Savings & Investments"
	DebtTransferParentName    = "Loan & Debt"
)

// BudgetActualsGetRequest represents the request for the unified budget actuals
type BudgetActualsGetRequest struct {
	StartTime int64 `form:"startTime" binding:"required,min=1"`
	EndTime   int64 `form:"endTime" binding:"required,min=1"`
}

// BudgetActualItem represents an actual amount for a single category within a single section
type BudgetActualItem struct {
	CategoryId int64  `json:"categoryId,string"`
	Section    string `json:"section"`
	Amount     int64  `json:"amount"`
}

// BudgetActualsResponse represents the unified budget actuals response
type BudgetActualsResponse struct {
	Items []*BudgetActualItem `json:"items"`
}

// BudgetLiabilitiesGetRequest represents the request for the budget liabilities reserve view
type BudgetLiabilitiesGetRequest struct {
	StartTime int64 `form:"startTime" binding:"required,min=1"`
	EndTime   int64 `form:"endTime" binding:"required,min=1"`
}

// BudgetLiabilityReserveItem represents the debt/reserve view for a single liability account
type BudgetLiabilityReserveItem struct {
	AccountId     int64  `json:"accountId,string"`
	Name          string `json:"name"`
	Icon          int64  `json:"icon,string"`
	Color         string `json:"color"`
	Currency      string `json:"currency"`
	Owed          int64  `json:"owed"`          // outstanding balance owed (positive = you owe)
	CycleSpend    int64  `json:"cycleSpend"`    // new charges on this account within the cycle
	CyclePayments int64  `json:"cyclePayments"` // payments into this account within the cycle
}

// BudgetLiabilitiesResponse represents the budget liabilities reserve response
type BudgetLiabilitiesResponse struct {
	Items []*BudgetLiabilityReserveItem `json:"items"`
}

// ToInfoResponse returns a view-object according to database model
func (b *BudgetTarget) ToInfoResponse() *BudgetTargetInfoResponse {
	return &BudgetTargetInfoResponse{
		Id:         b.Id,
		CategoryId: b.CategoryId,
		Section:    b.Section,
		Year:       b.Year,
		Month:      b.Month,
		Amount:     b.Amount,
	}
}
