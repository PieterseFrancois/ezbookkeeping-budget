package models

// Goal represents goal data stored in database
type Goal struct {
	Id           int64  `xorm:"PK"`
	Uid          int64  `xorm:"INDEX NOT NULL"`
	Name         string `xorm:"NOT NULL"`
	AccountId    int64  `xorm:"NOT NULL"`
	TargetAmount int64  `xorm:"NOT NULL"`
	TargetDate   int64  `xorm:"NOT NULL"`
	CreatedAt    int64  `xorm:"NOT NULL"`
}

// GoalCreateRequest represents all parameters of goal creation request
type GoalCreateRequest struct {
	Name         string `json:"name" binding:"required"`
	AccountId    int64  `json:"accountId,string" binding:"required,min=1"`
	TargetAmount int64  `json:"targetAmount,string" binding:"required,min=1"`
	TargetDate   int64  `json:"targetDate,string" binding:"required,min=1"`
}

// GoalUpdateRequest represents all parameters of goal modification request
type GoalUpdateRequest struct {
	Id           int64  `json:"id,string" binding:"required,min=1"`
	Name         string `json:"name" binding:"required"`
	AccountId    int64  `json:"accountId,string" binding:"required,min=1"`
	TargetAmount int64  `json:"targetAmount,string" binding:"required,min=1"`
	TargetDate   int64  `json:"targetDate,string" binding:"required,min=1"`
}

// GoalDeleteRequest represents all parameters of goal deleting request
type GoalDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// GoalInfoResponse represents a view-object of goal
type GoalInfoResponse struct {
	Id           int64  `json:"id,string"`
	Name         string `json:"name"`
	AccountId    int64  `json:"accountId,string"`
	TargetAmount int64  `json:"targetAmount,string"`
	TargetDate   int64  `json:"targetDate,string"`
	CreatedAt    int64  `json:"createdAt,string"`
}

// ToInfoResponse returns a view-object according to database model
func (g *Goal) ToInfoResponse() *GoalInfoResponse {
	return &GoalInfoResponse{
		Id:           g.Id,
		Name:         g.Name,
		AccountId:    g.AccountId,
		TargetAmount: g.TargetAmount,
		TargetDate:   g.TargetDate,
		CreatedAt:    g.CreatedAt,
	}
}
