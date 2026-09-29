package models

import (
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionDraft represents an unconfirmed transaction staged from an external integration or manual UI entry.
type TransactionDraft struct {
	TransactionDraftId   int64             `xorm:"PK"`
	Uid                  int64             `xorm:"UNIQUE(UQE_transaction_draft_uid_source) INDEX NOT NULL"`
	Type                 TransactionDbType `xorm:"NOT NULL"`
	CategoryId           int64
	AccountId            int64
	DestinationAccountId int64
	TransactionTime      int64 `xorm:"NOT NULL"`
	TimezoneUtcOffset    int16
	Amount               int64 `xorm:"NOT NULL"`
	DestinationAmount    int64
	HideAmount           bool
	TagIds               string `xorm:"VARCHAR(500)"`
	PictureIds           string `xorm:"VARCHAR(500)"`
	Comment              string `xorm:"VARCHAR(255)"`
	ExcludeFromBudget    bool
	Source               string `xorm:"UNIQUE(UQE_transaction_draft_uid_source) VARCHAR(255) NOT NULL"`
	CreatedIp            string `xorm:"VARCHAR(39)"`
	CreatedUnixTime      int64
	UpdatedUnixTime      int64
}

// TransactionDraftCreateRequest represents all parameters of transaction draft creation request
type TransactionDraftCreateRequest struct {
	Type                 TransactionType `json:"type" binding:"required"`
	CategoryId           int64           `json:"categoryId,string" binding:"min=0"`
	Time                 int64           `json:"time" binding:"required,min=1"`
	UtcOffset            int16           `json:"utcOffset" binding:"min=-720,max=840"`
	AccountId            int64           `json:"accountId,string" binding:"min=0"`
	DestinationAccountId int64           `json:"destinationAccountId,string" binding:"min=0"`
	Amount               int64           `json:"amount" binding:"min=-9999999999999,max=9999999999999"`
	DestinationAmount    int64           `json:"destinationAmount" binding:"min=-9999999999999,max=9999999999999"`
	HideAmount           bool            `json:"hideAmount"`
	TagIds               []string        `json:"tagIds"`
	PictureIds           []string        `json:"pictureIds"`
	Comment              string          `json:"comment" binding:"max=255"`
	ExcludeFromBudget    bool            `json:"excludeFromBudget"`
	Source               string          `json:"source" binding:"required,max=255"`
}

// TransactionDraftModifyRequest represents all parameters of transaction draft partial-update request.
// All business fields are optional (pointer types); nil/omitted means "don't change".
type TransactionDraftModifyRequest struct {
	Source               string           `json:"source" binding:"required,max=255"`
	Type                 *TransactionType `json:"type"`
	CategoryId           *int64           `json:"categoryId,string"`
	Time                 *int64           `json:"time"`
	UtcOffset            *int16           `json:"utcOffset"`
	AccountId            *int64           `json:"accountId,string"`
	DestinationAccountId *int64           `json:"destinationAccountId,string"`
	Amount               *int64           `json:"amount"`
	DestinationAmount    *int64           `json:"destinationAmount"`
	HideAmount           *bool            `json:"hideAmount"`
	TagIds               []string         `json:"tagIds"`
	PictureIds           []string         `json:"pictureIds"`
	Comment              *string          `json:"comment"`
	ExcludeFromBudget    *bool            `json:"excludeFromBudget"`
}

// TransactionDraftGetRequest represents all parameters of transaction draft (or confirmed transaction) getting-by-source request
type TransactionDraftGetRequest struct {
	Source string `json:"source" binding:"required,max=255"`
}

// TransactionDraftDeleteRequest represents all parameters of transaction draft (or confirmed transaction) deleting-by-source request
type TransactionDraftDeleteRequest struct {
	Source string `json:"source" binding:"required,max=255"`
}

// TransactionDraftConfirmRequest represents all parameters of transaction draft confirmation (promotion) request
type TransactionDraftConfirmRequest struct {
	Source               string `json:"source" binding:"required,max=255"`
	CategoryId           *int64 `json:"categoryId,string"`
	AccountId            *int64 `json:"accountId,string"`
	DestinationAccountId *int64 `json:"destinationAccountId,string"`
	ExcludeFromBudget    *bool  `json:"excludeFromBudget"`
}

// TransactionDraftCountResponse represents the total number of pending transaction drafts
type TransactionDraftCountResponse struct {
	TotalCount int64 `json:"totalCount"`
}

// TransactionDraftListRequest represents all parameters of transaction draft listing request
type TransactionDraftListRequest struct {
	Page  int32 `form:"page" binding:"min=0"`
	Count int32 `form:"count" binding:"required,min=1,max=50"`
}

// TransactionDraftInfoResponse represents a view-object of transaction draft
type TransactionDraftInfoResponse struct {
	Id                   int64                                    `json:"id,string"`
	Type                 TransactionType                          `json:"type"`
	CategoryId           int64                                    `json:"categoryId,string"`
	AccountId            int64                                    `json:"accountId,string"`
	DestinationAccountId int64                                    `json:"destinationAccountId,string,omitempty"`
	Time                 int64                                    `json:"time"`
	UtcOffset            int16                                    `json:"utcOffset"`
	Amount               int64                                    `json:"amount"`
	DestinationAmount    int64                                    `json:"destinationAmount,omitempty"`
	HideAmount           bool                                     `json:"hideAmount"`
	TagIds               []string                                 `json:"tagIds"`
	PictureIds           []string                                 `json:"pictureIds"`
	Pictures             TransactionPictureInfoBasicResponseSlice `json:"pictures,omitempty"`
	Comment              string                                   `json:"comment"`
	ExcludeFromBudget    bool                                     `json:"excludeFromBudget"`
	Source               string                                   `json:"source"`
	Complete             bool                                     `json:"complete"`
}

// ToTransactionDraftInfoResponse returns a view-object according to database model
func (t *TransactionDraft) ToTransactionDraftInfoResponse() *TransactionDraftInfoResponse {
	transactionType, err := t.Type.ToTransactionType()

	if err != nil {
		return nil
	}

	var tagIds []string

	if t.TagIds != "" {
		tagIds = splitCommaSeparatedIds(t.TagIds)
	} else {
		tagIds = make([]string, 0)
	}

	var pictureIds []string

	if t.PictureIds != "" {
		pictureIds = splitCommaSeparatedIds(t.PictureIds)
	} else {
		pictureIds = make([]string, 0)
	}

	return &TransactionDraftInfoResponse{
		Id:                   t.TransactionDraftId,
		Type:                 transactionType,
		CategoryId:           t.CategoryId,
		AccountId:            t.AccountId,
		DestinationAccountId: t.DestinationAccountId,
		Time:                 utils.GetUnixTimeFromTransactionTime(t.TransactionTime),
		UtcOffset:            t.TimezoneUtcOffset,
		Amount:               t.Amount,
		DestinationAmount:    t.DestinationAmount,
		HideAmount:           t.HideAmount,
		TagIds:               tagIds,
		PictureIds:           pictureIds,
		Comment:              t.Comment,
		ExcludeFromBudget:    t.ExcludeFromBudget,
		Source:               t.Source,
		Complete:             t.AccountId != 0 && t.CategoryId != 0,
	}
}

// TransactionOrDraftInfoResponse wraps either a confirmed transaction or a draft transaction, keyed by status
type TransactionOrDraftInfoResponse struct {
	Status      string                        `json:"status"`
	Transaction *TransactionInfoResponse      `json:"transaction,omitempty"`
	Draft       *TransactionDraftInfoResponse `json:"draft,omitempty"`
}

func splitCommaSeparatedIds(s string) []string {
	if s == "" {
		return []string{}
	}

	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}
