package services

import (
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// BudgetService represents budget target service
type BudgetService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a budget service singleton instance
var (
	BudgetTargets = &BudgetService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetSavingsActuals returns actual transfer amounts grouped by category for the given user, year and month.
// endDay controls the budget cycle boundary (0 = calendar month).
func (s *BudgetService) GetSavingsActuals(c core.Context, uid int64, year int, month int, endDay int, categoryIds []int64) ([]*models.SavingsCategoryActual, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	if len(categoryIds) == 0 {
		return []*models.SavingsCategoryActual{}, nil
	}

	minTransactionTime, maxTransactionTime, err := utils.GetTransactionTimeRangeByYearMonth(int32(year), int32(month), endDay)

	if err != nil {
		return nil, errs.ErrSystemError
	}

	var transactions []*models.Transaction
	err = s.UserDataDB(uid).NewSession(c).
		Select("transaction_id, category_id, amount, account_id").
		Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<=?",
			uid, false, minTransactionTime, maxTransactionTime).
		In("type", models.TRANSACTION_DB_TYPE_TRANSFER_OUT).
		In("category_id", categoryIds).
		Find(&transactions)

	if err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return []*models.SavingsCategoryActual{}, nil
	}

	// Filter out transactions excluded from budget calculations
	txIds := make([]int64, len(transactions))
	for i, t := range transactions {
		txIds[i] = t.TransactionId
	}

	excluded, err := TransactionBudgetOverrides.GetExcludedTransactionIds(c, uid, txIds)
	if err != nil {
		return nil, err
	}

	if len(excluded) > 0 {
		filtered := transactions[:0]
		for _, t := range transactions {
			if !excluded[t.TransactionId] {
				filtered = append(filtered, t)
			}
		}
		transactions = filtered

		if len(transactions) == 0 {
			return []*models.SavingsCategoryActual{}, nil
		}
	}

	// Collect unique source account IDs to classify contributions vs withdrawals
	accountIdSet := make(map[int64]struct{})
	for _, t := range transactions {
		accountIdSet[t.AccountId] = struct{}{}
	}
	accountIds := make([]int64, 0, len(accountIdSet))
	for id := range accountIdSet {
		accountIds = append(accountIds, id)
	}

	var accounts []*models.Account
	err = s.UserDataDB(uid).NewSession(c).
		Select("account_id, category").
		Where("uid=? AND deleted=?", uid, false).
		In("account_id", accountIds).
		Find(&accounts)

	if err != nil {
		return nil, err
	}

	// Mark accounts that are savings or investment accounts
	savingsAccountIds := make(map[int64]bool, len(accounts))
	for _, a := range accounts {
		if a.Category == models.ACCOUNT_CATEGORY_SAVINGS_ACCOUNT || a.Category == models.ACCOUNT_CATEGORY_INVESTMENT {
			savingsAccountIds[a.AccountId] = true
		}
	}

	resultMap := make(map[int64]*models.SavingsCategoryActual)

	for _, t := range transactions {
		actual, exists := resultMap[t.CategoryId]

		if !exists {
			actual = &models.SavingsCategoryActual{CategoryId: t.CategoryId}
			resultMap[t.CategoryId] = actual
		}

		// Source is a savings/investment account → money is leaving savings (withdrawal)
		// Source is any other account → money is flowing into savings (contribution)
		if savingsAccountIds[t.AccountId] {
			actual.TransferIn += t.Amount
		} else {
			actual.TransferOut += t.Amount
		}
	}

	result := make([]*models.SavingsCategoryActual, 0, len(resultMap))

	for _, actual := range resultMap {
		actual.Net = actual.TransferOut - actual.TransferIn
		result = append(result, actual)
	}

	return result, nil
}

// GetExpenseIncomeActuals returns actual expense and income amounts grouped by category for the given time range,
// filtered to exclude transactions marked as excluded from budget.
func (s *BudgetService) GetExpenseIncomeActuals(c core.Context, uid int64, startTime int64, endTime int64) ([]*models.BudgetCategoryActualItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(startTime)
	maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(endTime)

	var transactions []*models.Transaction
	err := s.UserDataDB(uid).NewSession(c).
		Select("transaction_id, category_id, amount").
		Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<=?",
			uid, false, minTransactionTime, maxTransactionTime).
		In("type", models.TRANSACTION_DB_TYPE_EXPENSE, models.TRANSACTION_DB_TYPE_INCOME).
		Find(&transactions)

	if err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return []*models.BudgetCategoryActualItem{}, nil
	}

	txIds := make([]int64, len(transactions))
	for i, t := range transactions {
		txIds[i] = t.TransactionId
	}

	excluded, err := TransactionBudgetOverrides.GetExcludedTransactionIds(c, uid, txIds)
	if err != nil {
		return nil, err
	}

	totals := make(map[int64]int64)
	for _, t := range transactions {
		if excluded[t.TransactionId] {
			continue
		}
		totals[t.CategoryId] += t.Amount
	}

	result := make([]*models.BudgetCategoryActualItem, 0, len(totals))
	for catId, amount := range totals {
		result = append(result, &models.BudgetCategoryActualItem{CategoryId: catId, Amount: amount})
	}

	return result, nil
}

// GetBudgetTargets returns all budget targets for the given user, year and month
func (s *BudgetService) GetBudgetTargets(c core.Context, uid int64, year int, month int) ([]*models.BudgetTarget, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var targets []*models.BudgetTarget
	err := s.UserDataDB(uid).NewSession(c).Where("uid=? AND year=? AND month=?", uid, year, month).Find(&targets)

	return targets, err
}

// CreateBudgetTarget saves a new budget target model to database
func (s *BudgetService) CreateBudgetTarget(c core.Context, uid int64, request *models.BudgetTargetCreateRequest) (*models.BudgetTarget, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	exists, err := s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND category_id=? AND year=? AND month=?", uid, request.CategoryId, request.Year, request.Month).
		Exist(&models.BudgetTarget{})

	if err != nil {
		return nil, err
	} else if exists {
		return nil, errs.ErrBudgetTargetAlreadyExists
	}

	target := &models.BudgetTarget{
		Id:         s.GenerateUuid(uuid.UUID_TYPE_BUDGET),
		Uid:        uid,
		CategoryId: request.CategoryId,
		Year:       request.Year,
		Month:      request.Month,
		Amount:     request.Amount,
	}

	if target.Id < 1 {
		return nil, errs.ErrSystemIsBusy
	}

	err = s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(target)
		return err
	})

	if err != nil {
		return nil, err
	}

	return target, nil
}

// UpdateBudgetTarget saves an existed budget target model to database
func (s *BudgetService) UpdateBudgetTarget(c core.Context, uid int64, request *models.BudgetTargetUpdateRequest) (*models.BudgetTarget, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	target := &models.BudgetTarget{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(request.Id).Where("uid=?", uid).Get(target)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrBudgetTargetNotFound
	}

	target.Amount = request.Amount

	err = s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(target.Id).Cols("amount").Where("uid=?", uid).Update(target)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrBudgetTargetNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return target, nil
}

// DeleteBudgetTarget deletes an existed budget target from database
func (s *BudgetService) DeleteBudgetTarget(c core.Context, uid int64, id int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return errs.ErrBudgetTargetIdInvalid
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(id).Where("uid=?", uid).Delete(&models.BudgetTarget{})

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrBudgetTargetNotFound
		}

		return nil
	})
}
