package models

// TransactionBudgetOverride marks a transaction with budget calculation overrides.
// Currently supports excluding a transaction from budget actuals.
// Future columns (OverrideType, OverrideCategoryId) can be added via SyncStructs without data loss.
type TransactionBudgetOverride struct {
	OverrideId      int64 `xorm:"PK AUTOINCR"`
	Uid             int64 `xorm:"UNIQUE(UQE_tx_budget_override_uid_tx) NOT NULL"`
	TransactionId   int64 `xorm:"UNIQUE(UQE_tx_budget_override_uid_tx) NOT NULL"`
	Excluded        bool  `xorm:"NOT NULL DEFAULT 0"`
	CreatedUnixTime int64
	UpdatedUnixTime int64
}

// TransactionBudgetOverrideSetRequest represents the payload for setting a budget override
type TransactionBudgetOverrideSetRequest struct {
	TransactionId int64 `json:"transactionId,string" binding:"required,min=1"`
	Excluded      bool  `json:"excluded"`
}
