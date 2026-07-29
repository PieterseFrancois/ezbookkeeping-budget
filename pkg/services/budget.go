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

// GetLiabilityReserves returns the debt/reserve view per liability account for the given cycle:
// what is owed (from the account balance) plus new charges and payments within the cycle.
// This is a factual cash-flow view of debt, so budget exclusions are intentionally not applied here.
func (s *BudgetService) GetLiabilityReserves(c core.Context, uid int64, startTime int64, endTime int64) ([]*models.BudgetLiabilityReserveItem, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var accounts []*models.Account
	err := s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND deleted=?", uid, false).
		Find(&accounts)

	if err != nil {
		return nil, err
	}

	// Keep leaf liability accounts (exclude multi-sub parent accounts, whose sub-accounts carry the balances/transactions)
	reserves := make(map[int64]*models.BudgetLiabilityReserveItem)
	accountIds := make([]int64, 0)

	for _, a := range accounts {
		if !a.Category.IsLiability() || a.Type == models.ACCOUNT_TYPE_MULTI_SUB_ACCOUNTS {
			continue
		}

		reserves[a.AccountId] = &models.BudgetLiabilityReserveItem{
			AccountId: a.AccountId,
			Name:      a.Name,
			Icon:      a.Icon,
			Color:     a.Color,
			Currency:  a.Currency,
			Owed:      -a.Balance, // liability balances are stored negative; owed magnitude is the negation
		}
		accountIds = append(accountIds, a.AccountId)
	}

	if len(accountIds) == 0 {
		return []*models.BudgetLiabilityReserveItem{}, nil
	}

	minTransactionTime := utils.GetMinTransactionTimeFromUnixTime(startTime)
	maxTransactionTime := utils.GetMaxTransactionTimeFromUnixTime(endTime)

	var transactions []*models.Transaction
	err = s.UserDataDB(uid).NewSession(c).
		Select("type, account_id, amount").
		Where("uid=? AND deleted=? AND transaction_time>=? AND transaction_time<=?",
			uid, false, minTransactionTime, maxTransactionTime).
		In("account_id", accountIds).
		In("type", models.TRANSACTION_DB_TYPE_EXPENSE, models.TRANSACTION_DB_TYPE_TRANSFER_IN).
		Find(&transactions)

	if err != nil {
		return nil, err
	}

	for _, t := range transactions {
		r := reserves[t.AccountId]

		if r == nil {
			continue
		}

		switch t.Type {
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			r.CycleSpend += t.Amount // a purchase charged to the card increases what is owed
		case models.TRANSACTION_DB_TYPE_TRANSFER_IN:
			r.CyclePayments += t.Amount // a transfer into the card is a repayment
		}
	}

	result := make([]*models.BudgetLiabilityReserveItem, 0, len(accountIds))

	for _, id := range accountIds {
		result = append(result, reserves[id])
	}

	return result, nil
}

// naturalSectionForCategory returns the section a budget target belongs to when none was specified.
// Expense/income categories map by type; transfer categories map by their parent category name.
func naturalSectionForCategory(category *models.TransactionCategory, parentName string) string {
	switch category.Type {
	case models.CATEGORY_TYPE_EXPENSE:
		return models.BUDGET_SECTION_EXPENSE
	case models.CATEGORY_TYPE_INCOME:
		return models.BUDGET_SECTION_INCOME
	case models.CATEGORY_TYPE_TRANSFER:
		if parentName == models.DebtTransferParentName {
			return models.BUDGET_SECTION_DEBT
		}
		// Savings & Investments, and anything else transfer-shaped, is a contribution
		return models.BUDGET_SECTION_SAVINGS
	default:
		return models.BUDGET_SECTION_EXPENSE
	}
}

// resolveNaturalSection looks up a single category and returns the section it naturally belongs to.
func (s *BudgetService) resolveNaturalSection(c core.Context, uid int64, categoryId int64) (string, error) {
	category := &models.TransactionCategory{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(categoryId).Where("uid=?", uid).Get(category)

	if err != nil {
		return "", err
	} else if !has {
		return "", errs.ErrTransactionCategoryNotFound
	}

	parentName := ""

	if category.ParentCategoryId > models.LevelOneTransactionCategoryParentId {
		parent := &models.TransactionCategory{}
		hasParent, err := s.UserDataDB(uid).NewSession(c).ID(category.ParentCategoryId).Where("uid=?", uid).Get(parent)

		if err != nil {
			return "", err
		} else if hasParent {
			parentName = parent.Name
		}
	}

	return naturalSectionForCategory(category, parentName), nil
}

// BackfillBudgetTargetSections fills in the section on budget targets created before sections existed.
// It only touches rows with an empty section, so it is idempotent and safe to run on every startup.
// Runs across every user-data shard, so it needs no user enumeration. Returns the number of rows updated.
func (s *BudgetService) BackfillBudgetTargetSections(c core.Context) (int, error) {
	totalUpdated := 0

	for i := 0; i < s.UserDataDBCount(); i++ {
		db := s.UserDataDBByIndex(i)

		var targets []*models.BudgetTarget
		err := db.NewSession(c).Where("section=? OR section IS NULL", "").Find(&targets)

		if err != nil {
			return totalUpdated, err
		}

		if len(targets) == 0 {
			continue
		}

		var categories []*models.TransactionCategory
		err = db.NewSession(c).Find(&categories)

		if err != nil {
			return totalUpdated, err
		}

		categoryMap := make(map[int64]*models.TransactionCategory, len(categories))
		for _, cat := range categories {
			categoryMap[cat.CategoryId] = cat
		}

		err = db.DoTransaction(c, func(sess *xorm.Session) error {
			for _, target := range targets {
				category := categoryMap[target.CategoryId]

				if category == nil {
					// Category no longer exists; default to expense so the row is never left blank
					target.Section = models.BUDGET_SECTION_EXPENSE
				} else {
					parentName := ""

					if parent := categoryMap[category.ParentCategoryId]; parent != nil {
						parentName = parent.Name
					}

					target.Section = naturalSectionForCategory(category, parentName)
				}

				if _, err := sess.ID(target.Id).Cols("section").Update(target); err != nil {
					return err
				}

				totalUpdated++
			}

			return nil
		})

		if err != nil {
			return totalUpdated, err
		}
	}

	return totalUpdated, nil
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

	section := request.Section

	if section == "" {
		// Older clients omit the section; resolve the category's natural one so the row is never left blank
		natural, err := s.resolveNaturalSection(c, uid, request.CategoryId)

		if err != nil {
			return nil, err
		}

		section = natural
	}

	exists, err := s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND category_id=? AND section=? AND year=? AND month=?", uid, request.CategoryId, section, request.Year, request.Month).
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
		Section:    section,
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
