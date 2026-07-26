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

// isSavingsOrInvestmentCategory reports whether an account category holds savings/investments (i.e. "net-worth" money).
func isSavingsOrInvestmentCategory(cat models.AccountCategory) bool {
	return cat == models.ACCOUNT_CATEGORY_SAVINGS_ACCOUNT || cat == models.ACCOUNT_CATEGORY_INVESTMENT
}

// isSpendableCategory reports whether an account category is spendable cash (an asset that is not savings/investment).
func isSpendableCategory(cat models.AccountCategory) bool {
	return cat.IsAsset() && !isSavingsOrInvestmentCategory(cat)
}

// classifyTransferSection maps a transfer (by its source and destination account categories) to a budget section.
// Returns "" for transfers that are not budgeted (e.g. moving cash between two spendable accounts).
func classifyTransferSection(src models.AccountCategory, dst models.AccountCategory) string {
	if isSpendableCategory(src) && isSavingsOrInvestmentCategory(dst) {
		return models.BUDGET_SECTION_SAVINGS // spendable -> savings/investment (a contribution)
	}

	if isSpendableCategory(src) && dst.IsLiability() {
		return models.BUDGET_SECTION_DEBT // spendable -> liability (a card/loan paydown)
	}

	if isSavingsOrInvestmentCategory(src) && isSpendableCategory(dst) {
		return models.BUDGET_SECTION_INCOME // savings/investment -> spendable (a withdrawal, acts like income)
	}

	return ""
}

// GetBudgetActuals returns actual amounts grouped by (category, section) for the given time range, in a single pass.
// Expense/income transactions map to their section by type; transfers map by the direction of money relative to
// spendable cash (see classifyTransferSection). Transactions excluded from budget are skipped across all sections.
func (s *BudgetService) GetBudgetActuals(c core.Context, uid int64, startTime int64, endTime int64) ([]*models.BudgetActualItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(startTime)
	maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(endTime)

	var transactions []*models.Transaction
	err := s.UserDataDB(uid).NewSession(c).
		Select("transaction_id, type, category_id, amount, account_id, related_account_id").
		Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<=?",
			uid, false, minTransactionTime, maxTransactionTime).
		In("type", models.TRANSACTION_DB_TYPE_EXPENSE, models.TRANSACTION_DB_TYPE_INCOME, models.TRANSACTION_DB_TYPE_TRANSFER_OUT).
		Find(&transactions)

	if err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return []*models.BudgetActualItem{}, nil
	}

	// Filter out transactions excluded from budget calculations (applies to every section)
	txIds := make([]int64, len(transactions))
	for i, t := range transactions {
		txIds[i] = t.TransactionId
	}

	excluded, err := TransactionBudgetOverrides.GetExcludedTransactionIds(c, uid, txIds)
	if err != nil {
		return nil, err
	}

	// Collect account ids referenced by transfers so we can classify each leg by account category
	accountIdSet := make(map[int64]struct{})
	for _, t := range transactions {
		if t.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT {
			accountIdSet[t.AccountId] = struct{}{}
			accountIdSet[t.RelatedAccountId] = struct{}{}
		}
	}

	accountCategory := make(map[int64]models.AccountCategory)

	if len(accountIdSet) > 0 {
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

		for _, a := range accounts {
			accountCategory[a.AccountId] = a.Category
		}
	}

	// Aggregate amounts by (categoryId, section)
	type sectionKey struct {
		categoryId int64
		section    string
	}

	totals := make(map[sectionKey]int64)

	for _, t := range transactions {
		if excluded[t.TransactionId] {
			continue
		}

		var section string

		switch t.Type {
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			section = models.BUDGET_SECTION_EXPENSE
		case models.TRANSACTION_DB_TYPE_INCOME:
			section = models.BUDGET_SECTION_INCOME
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			section = classifyTransferSection(accountCategory[t.AccountId], accountCategory[t.RelatedAccountId])
		}

		if section == "" {
			continue
		}

		totals[sectionKey{categoryId: t.CategoryId, section: section}] += t.Amount
	}

	result := make([]*models.BudgetActualItem, 0, len(totals))

	for key, amount := range totals {
		result = append(result, &models.BudgetActualItem{
			CategoryId: key.categoryId,
			Section:    key.section,
			Amount:     amount,
		})
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
