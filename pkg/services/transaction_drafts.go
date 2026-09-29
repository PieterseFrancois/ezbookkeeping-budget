package services

import (
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// TransactionDraftService represents transaction draft service
type TransactionDraftService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// TransactionDrafts is the singleton service instance
var TransactionDrafts = &TransactionDraftService{
	ServiceUsingDB: ServiceUsingDB{
		container: datastore.Container,
	},
	ServiceUsingUuid: ServiceUsingUuid{
		container: uuid.Container,
	},
}

// isDuplicateKeyError returns whether the given error represents a unique-constraint violation
// on any of the database engines this project supports (MySQL, Postgres, SQLite3). It is written
// defensively: an unrecognized error type simply falls through to returning false rather than
// panicking, so it fails closed if a driver isn't actually compiled in for some build-tag reason.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	var mysqlErr *mysql.MySQLError

	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}

	var pqErr *pq.Error

	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}

	var sqliteErr sqlite3.Error

	if errors.As(err, &sqliteErr) {
		return sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique || sqliteErr.Code == sqlite3.ErrConstraint
	}

	// Fallback for any driver error type not recognized above
	return strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

// CreateTransactionDraft saves a new transaction draft to database
func (s *TransactionDraftService) CreateTransactionDraft(c core.Context, draft *models.TransactionDraft) error {
	if draft.Uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if draft.Source == "" {
		return errs.ErrTransactionDraftSourceIsEmpty
	}

	// The draft table's unique index only protects against a draft/draft race on this source; a
	// confirmed transaction lives in a different table with its own independent unique index, so
	// without this check a source already claimed by an active confirmed transaction could end up
	// held by a draft too. This is a pre-check (not DB-enforced across tables) so a very tight race
	// against a concurrent confirm is still possible in principle, but the realistic race this
	// guards against - re-creating a draft for a source that was already confirmed - isn't a
	// webhook-retry timing race, so a pre-check is an acceptable trade-off here.
	if _, err := Transactions.GetTransactionBySource(c, draft.Uid, draft.Source); err == nil {
		return errs.ErrTransactionDraftSourceAlreadyExists
	} else if err != errs.ErrTransactionNotFound {
		return err
	}

	transactionDraftUuids := s.GenerateUuids(uuid.UUID_TYPE_TRANSACTION, 1)

	if len(transactionDraftUuids) < 1 {
		return errs.ErrSystemIsBusy
	}

	now := time.Now().Unix()

	draft.TransactionDraftId = transactionDraftUuids[0]
	draft.CreatedUnixTime = now
	draft.UpdatedUnixTime = now

	err := s.UserDataDB(draft.Uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(draft)
		return err
	})

	if err != nil {
		if isDuplicateKeyError(err) {
			return errs.ErrTransactionDraftSourceAlreadyExists
		}

		return err
	}

	return nil
}

// GetTransactionDraftBySource returns a transaction draft model according to its source
func (s *TransactionDraftService) GetTransactionDraftBySource(c core.Context, uid int64, source string) (*models.TransactionDraft, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if source == "" {
		return nil, errs.ErrTransactionDraftSourceIsEmpty
	}

	draft := &models.TransactionDraft{}
	has, err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND source=?", uid, source).Get(draft)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrTransactionDraftNotFound
	}

	return draft, nil
}

// ListTransactionDraftsByUser returns paginated transaction drafts for a user, newest first
func (s *TransactionDraftService) ListTransactionDraftsByUser(c core.Context, uid int64, page int32, count int32) ([]*models.TransactionDraft, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if page < 0 {
		page = 0
	}

	if count <= 0 {
		count = 1
	}

	var drafts []*models.TransactionDraft
	err := s.UserDataDB(uid).NewSession(c).
		Where("uid=?", uid).
		OrderBy("transaction_draft_id desc").
		Limit(int(count), int(page)*int(count)).
		Find(&drafts)

	if err != nil {
		return nil, err
	}

	return drafts, nil
}

// GetTransactionDraftCount returns the total number of pending transaction drafts for a user
func (s *TransactionDraftService) GetTransactionDraftCount(c core.Context, uid int64) (int64, error) {
	if uid <= 0 {
		return 0, errs.ErrUserIdInvalid
	}

	return s.UserDataDB(uid).NewSession(c).Where("uid=?", uid).Count(&models.TransactionDraft{})
}

// ModifyTransactionDraftBySource applies a partial update to a transaction draft, identified by its source
func (s *TransactionDraftService) ModifyTransactionDraftBySource(c core.Context, uid int64, source string, updates *models.TransactionDraftModifyRequest) (*models.TransactionDraft, error) {
	draft, err := s.GetTransactionDraftBySource(c, uid, source)

	if err != nil {
		return nil, err
	}

	updateCols := make([]string, 0, 16)

	if updates.Type != nil {
		dbType, err := (*updates.Type).ToTransactionDbType()

		if err != nil {
			return nil, err
		}

		draft.Type = dbType
		updateCols = append(updateCols, "type")
	}

	if updates.CategoryId != nil {
		draft.CategoryId = *updates.CategoryId
		updateCols = append(updateCols, "category_id")
	}

	if updates.Time != nil {
		draft.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(*updates.Time)
		updateCols = append(updateCols, "transaction_time")
	}

	if updates.UtcOffset != nil {
		draft.TimezoneUtcOffset = *updates.UtcOffset
		updateCols = append(updateCols, "timezone_utc_offset")
	}

	if updates.AccountId != nil {
		draft.AccountId = *updates.AccountId
		updateCols = append(updateCols, "account_id")
	}

	if updates.DestinationAccountId != nil {
		draft.DestinationAccountId = *updates.DestinationAccountId
		updateCols = append(updateCols, "destination_account_id")
	}

	if updates.Amount != nil {
		draft.Amount = *updates.Amount
		updateCols = append(updateCols, "amount")
	}

	if updates.DestinationAmount != nil {
		draft.DestinationAmount = *updates.DestinationAmount
		updateCols = append(updateCols, "destination_amount")
	}

	if updates.HideAmount != nil {
		draft.HideAmount = *updates.HideAmount
		updateCols = append(updateCols, "hide_amount")
	}

	if updates.TagIds != nil {
		tagIds, err := utils.StringArrayToInt64Array(updates.TagIds)

		if err != nil {
			return nil, errs.ErrTransactionTagIdInvalid
		}

		draft.TagIds = strings.Join(utils.Int64ArrayToStringArray(utils.ToUniqueInt64Slice(tagIds)), ",")
		updateCols = append(updateCols, "tag_ids")
	}

	if updates.PictureIds != nil {
		pictureIds, err := utils.StringArrayToInt64Array(updates.PictureIds)

		if err != nil {
			return nil, errs.ErrTransactionPictureIdInvalid
		}

		draft.PictureIds = strings.Join(utils.Int64ArrayToStringArray(utils.ToUniqueInt64Slice(pictureIds)), ",")
		updateCols = append(updateCols, "picture_ids")
	}

	if updates.Comment != nil {
		draft.Comment = *updates.Comment
		updateCols = append(updateCols, "comment")
	}

	if updates.ExcludeFromBudget != nil {
		draft.ExcludeFromBudget = *updates.ExcludeFromBudget
		updateCols = append(updateCols, "exclude_from_budget")
	}

	if len(updateCols) == 0 {
		return draft, nil
	}

	draft.UpdatedUnixTime = time.Now().Unix()
	updateCols = append(updateCols, "updated_unix_time")

	err = s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.ID(draft.TransactionDraftId).Cols(updateCols...).Update(draft)
		return err
	})

	if err != nil {
		return nil, err
	}

	return draft, nil
}

// DeleteTransactionDraftBySource permanently deletes a transaction draft, identified by its source
func (s *TransactionDraftService) DeleteTransactionDraftBySource(c core.Context, uid int64, source string) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if source == "" {
		return errs.ErrTransactionDraftSourceIsEmpty
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Where("uid=? AND source=?", uid, source).Delete(&models.TransactionDraft{})
		return err
	})
}

// ConfirmTransactionDraft promotes a transaction draft into a real, confirmed transaction.
// The draft is only deleted after the confirmed transaction has been successfully created
// (create-then-delete, never delete-then-create), so a failure at any point leaves either the
// original draft intact, or the confirmed transaction correctly created with, at worst, a harmless
// leftover draft row (which GetTransactionOrDraftBySource tolerates, since it prefers a draft match
// only when one still exists).
func (s *TransactionDraftService) ConfirmTransactionDraft(c core.Context, uid int64, source string, overrides *models.TransactionDraftConfirmRequest) (*models.Transaction, error) {
	draft, err := s.GetTransactionDraftBySource(c, uid, source)

	if err != nil {
		return nil, err
	}

	if overrides != nil {
		if overrides.CategoryId != nil {
			draft.CategoryId = *overrides.CategoryId
		}

		if overrides.AccountId != nil {
			draft.AccountId = *overrides.AccountId
		}

		if overrides.DestinationAccountId != nil {
			draft.DestinationAccountId = *overrides.DestinationAccountId
		}

		if overrides.ExcludeFromBudget != nil {
			draft.ExcludeFromBudget = *overrides.ExcludeFromBudget
		}
	}

	// A modify-balance draft has no category by definition (mirrors TransactionCreateHandler's
	// treatment of TRANSACTION_DB_TYPE_MODIFY_BALANCE, which forbids setting a category at all).
	if draft.AccountId == 0 || (draft.CategoryId == 0 && draft.Type != models.TRANSACTION_DB_TYPE_MODIFY_BALANCE) {
		return nil, errs.ErrTransactionDraftIncomplete
	}

	tagIds, err := utils.StringArrayToInt64Array(splitNonEmpty(draft.TagIds))

	if err != nil {
		return nil, errs.ErrTransactionTagIdInvalid
	}

	pictureIds, err := utils.StringArrayToInt64Array(splitNonEmpty(draft.PictureIds))

	if err != nil {
		return nil, errs.ErrTransactionPictureIdInvalid
	}

	draftSource := draft.Source
	transaction := &models.Transaction{
		TransactionId:     draft.TransactionDraftId,
		Uid:               uid,
		Type:              draft.Type,
		CategoryId:        draft.CategoryId,
		AccountId:         draft.AccountId,
		TransactionTime:   draft.TransactionTime,
		TimezoneUtcOffset: draft.TimezoneUtcOffset,
		Amount:            draft.Amount,
		HideAmount:        draft.HideAmount,
		Comment:           draft.Comment,
		CreatedIp:         draft.CreatedIp,
		Source:            &draftSource,
	}

	if draft.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT || draft.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		transaction.RelatedAccountId = draft.DestinationAccountId
		transaction.RelatedAccountAmount = draft.DestinationAmount
	}

	// A previously-confirmed transaction for this same source may have since been (soft-)deleted;
	// its row still holds the source value and would otherwise collide with the unique index below.
	if releaseErr := Transactions.ReleaseSourceFromDeletedTransactions(c, uid, draftSource); releaseErr != nil {
		log.Errorf(c, "[transaction_drafts.ConfirmTransactionDraft] failed to release source from deleted transactions for user \"uid:%d\", because %s", uid, releaseErr.Error())
		return nil, releaseErr
	}

	err = s.confirmCreateTransaction(c, transaction, tagIds, pictureIds)

	if err != nil {
		return nil, err
	}

	if draft.ExcludeFromBudget {
		if excludeErr := TransactionBudgetOverrides.SetExclusion(c, uid, transaction.TransactionId, true); excludeErr != nil {
			log.Errorf(c, "[transaction_drafts.ConfirmTransactionDraft] failed to set budget exclusion for transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, excludeErr.Error())
		}
	}

	if deleteErr := s.DeleteTransactionDraftBySource(c, uid, source); deleteErr != nil {
		log.Errorf(c, "[transaction_drafts.ConfirmTransactionDraft] failed to delete confirmed draft \"source:%s\" for user \"uid:%d\", because %s", source, uid, deleteErr.Error())
	}

	return transaction, nil
}

// confirmCreateTransaction calls the existing, unmodified TransactionService.CreateTransaction to
// persist the promoted transaction. It is a thin wrapper kept only so this file doesn't need to
// import Transactions at every call site.
func (s *TransactionDraftService) confirmCreateTransaction(c core.Context, transaction *models.Transaction, tagIds []int64, pictureIds []int64) error {
	return Transactions.CreateTransaction(c, transaction, tagIds, pictureIds)
}

// GetTransactionOrDraftBySource looks up either a draft or a confirmed transaction by source,
// preferring the draft when both somehow exist (which should not normally happen).
func (s *TransactionDraftService) GetTransactionOrDraftBySource(c core.Context, uid int64, source string) (*models.TransactionOrDraftInfoResponse, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if source == "" {
		return nil, errs.ErrTransactionDraftSourceIsEmpty
	}

	draft, err := s.GetTransactionDraftBySource(c, uid, source)

	if err == nil {
		return &models.TransactionOrDraftInfoResponse{
			Status: "draft",
			Draft:  draft.ToTransactionDraftInfoResponse(),
		}, nil
	} else if err != errs.ErrTransactionDraftNotFound {
		return nil, err
	}

	transaction, err := Transactions.GetTransactionBySource(c, uid, source)

	if err == nil {
		tagIdsMap, tagErr := TransactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})
		var tagIds []int64

		if tagErr == nil {
			tagIds = tagIdsMap[transaction.TransactionId]
		}

		editable := true

		return &models.TransactionOrDraftInfoResponse{
			Status:      "confirmed",
			Transaction: transaction.ToTransactionInfoResponse(tagIds, editable),
		}, nil
	} else if err != errs.ErrTransactionNotFound {
		return nil, err
	}

	return nil, errs.ErrTransactionDraftNotFound
}

// splitNonEmpty splits a comma-separated id string into a slice, returning an empty (non-nil) slice
// when the input is empty, mirroring how TagIds/PictureIds are represented on request DTOs elsewhere.
func splitNonEmpty(s string) []string {
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
