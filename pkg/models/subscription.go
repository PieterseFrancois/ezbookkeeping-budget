package models

// SubscriptionFrequencyType represents subscription billing frequency
type SubscriptionFrequencyType int8

// Subscription frequency types
const (
	SUBSCRIPTION_FREQUENCY_TYPE_DAILY     SubscriptionFrequencyType = 1
	SUBSCRIPTION_FREQUENCY_TYPE_WEEKLY    SubscriptionFrequencyType = 2
	SUBSCRIPTION_FREQUENCY_TYPE_MONTHLY   SubscriptionFrequencyType = 3
	SUBSCRIPTION_FREQUENCY_TYPE_QUARTERLY SubscriptionFrequencyType = 4
	SUBSCRIPTION_FREQUENCY_TYPE_ANNUAL    SubscriptionFrequencyType = 5
)

// Subscription represents subscription data stored in database
type Subscription struct {
	Id         int64                     `xorm:"PK"`
	Uid        int64                     `xorm:"INDEX NOT NULL"`
	Name       string                    `xorm:"VARCHAR(64) NOT NULL"`
	Amount     int64                     `xorm:"NOT NULL"`
	Currency   string                    `xorm:"VARCHAR(3) NOT NULL"`
	CategoryId int64                     `xorm:"NOT NULL"`
	TemplateId int64                     `xorm:"NOT NULL"`
	StartDate  int64                     `xorm:"NOT NULL"`
	Frequency  SubscriptionFrequencyType `xorm:"NOT NULL"`
	IsActive   bool                      `xorm:"NOT NULL"`
	CreatedAt  int64                     `xorm:"NOT NULL"`
}

// SubscriptionCreateRequest represents all parameters of subscription creation request
type SubscriptionCreateRequest struct {
	Name       string                    `json:"name" binding:"required,notBlank,max=64"`
	Amount     int64                     `json:"amount,string" binding:"required,min=1"`
	Currency   string                    `json:"currency" binding:"required,len=3,validCurrency"`
	CategoryId int64                     `json:"categoryId,string" binding:"min=0"`
	TemplateId int64                     `json:"templateId,string" binding:"min=0"`
	StartDate  int64                     `json:"startDate,string" binding:"required,min=1"`
	Frequency  SubscriptionFrequencyType `json:"frequency" binding:"required"`
	IsActive   bool                      `json:"isActive"`
}

// SubscriptionUpdateRequest represents all parameters of subscription modification request
type SubscriptionUpdateRequest struct {
	Id         int64                     `json:"id,string" binding:"required,min=1"`
	Name       string                    `json:"name" binding:"required,notBlank,max=64"`
	Amount     int64                     `json:"amount,string" binding:"required,min=1"`
	Currency   string                    `json:"currency" binding:"required,len=3,validCurrency"`
	CategoryId int64                     `json:"categoryId,string" binding:"min=0"`
	TemplateId int64                     `json:"templateId,string" binding:"min=0"`
	StartDate  int64                     `json:"startDate,string" binding:"required,min=1"`
	Frequency  SubscriptionFrequencyType `json:"frequency" binding:"required"`
	IsActive   bool                      `json:"isActive"`
}

// SubscriptionDeleteRequest represents all parameters of subscription deleting request
type SubscriptionDeleteRequest struct {
	Id int64 `json:"id,string" binding:"required,min=1"`
}

// SubscriptionToggleActiveRequest represents all parameters of subscription active state toggling request
type SubscriptionToggleActiveRequest struct {
	Id       int64 `json:"id,string" binding:"required,min=1"`
	IsActive bool  `json:"isActive"`
}

// SubscriptionInfoResponse represents a view-object of subscription
type SubscriptionInfoResponse struct {
	Id         int64                     `json:"id,string"`
	Name       string                    `json:"name"`
	Amount     int64                     `json:"amount,string"`
	Currency   string                    `json:"currency"`
	CategoryId int64                     `json:"categoryId,string"`
	TemplateId int64                     `json:"templateId,string"`
	StartDate  int64                     `json:"startDate,string"`
	Frequency  SubscriptionFrequencyType `json:"frequency"`
	IsActive   bool                      `json:"isActive"`
	CreatedAt  int64                     `json:"createdAt,string"`
}

// ToInfoResponse returns a view-object according to database model
func (s *Subscription) ToInfoResponse() *SubscriptionInfoResponse {
	return &SubscriptionInfoResponse{
		Id:         s.Id,
		Name:       s.Name,
		Amount:     s.Amount,
		Currency:   s.Currency,
		CategoryId: s.CategoryId,
		TemplateId: s.TemplateId,
		StartDate:  s.StartDate,
		Frequency:  s.Frequency,
		IsActive:   s.IsActive,
		CreatedAt:  s.CreatedAt,
	}
}
