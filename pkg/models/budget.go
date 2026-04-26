package models

// BudgetTarget represents budget target data stored in database
type BudgetTarget struct {
	Id         int64 `xorm:"PK"`
	Uid        int64 `xorm:"UNIQUE(IDX_budget_uid_category_year_month) NOT NULL"`
	CategoryId int64 `xorm:"UNIQUE(IDX_budget_uid_category_year_month) NOT NULL"`
	Year       int   `xorm:"UNIQUE(IDX_budget_uid_category_year_month) NOT NULL"`
	Month      int   `xorm:"UNIQUE(IDX_budget_uid_category_year_month) NOT NULL"`
	Amount     int64 `xorm:"NOT NULL"`
}

// BudgetTargetCreateRequest represents all parameters of budget target creation request
type BudgetTargetCreateRequest struct {
	CategoryId int64 `json:"categoryId,string" binding:"required,min=1"`
	Year       int   `json:"year" binding:"required,min=1"`
	Month      int   `json:"month" binding:"required,min=1,max=12"`
	Amount	   int64 `json:"amount,string" binding:"min=0"`
}

// BudgetTargetUpdateRequest represents all parameters of budget target modification request
type BudgetTargetUpdateRequest struct {
	Id     int64 `json:"id,string" binding:"required,min=1"`
	Amount int64 `json:"amount,string" binding:"min=0"`
}

// BudgetTargetInfoResponse represents a view-object of budget target
type BudgetTargetInfoResponse struct {
	Id         int64 `json:"id,string"`
	CategoryId int64 `json:"categoryId,string"`
	Year       int   `json:"year"`
	Month      int   `json:"month"`
	Amount     int64 `json:"amount,string"`
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

// ToInfoResponse returns a view-object according to database model
func (b *BudgetTarget) ToInfoResponse() *BudgetTargetInfoResponse {
	return &BudgetTargetInfoResponse{
		Id:         b.Id,
		CategoryId: b.CategoryId,
		Year:       b.Year,
		Month:      b.Month,
		Amount:     b.Amount,
	}
}
