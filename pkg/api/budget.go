package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// BudgetApi represents budget api
type BudgetApi struct {
	budgetTargets            *services.BudgetService
	transactionCategories    *services.TransactionCategoryService
	transactionBudgetOverrides *services.TransactionBudgetOverrideService
	users                    *services.UserService
}

// Initialize a budget api singleton instance
var (
	Budget = &BudgetApi{
		budgetTargets:              services.BudgetTargets,
		transactionCategories:      services.TransactionCategories,
		transactionBudgetOverrides: services.TransactionBudgetOverrides,
		users:                      services.Users,
	}
)

// SavingsActualsHandler returns actual transfer amounts per transfer subcategory for the given year and month
func (a *BudgetApi) SavingsActualsHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.SavingsActualGetRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[budget.SavingsActualsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		log.Errorf(c, "[budget.SavingsActualsHandler] failed to get user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrUserNotFound)
	}

	allCategories, err := a.transactionCategories.GetAllCategoriesByUid(c, uid, models.CATEGORY_TYPE_TRANSFER, -1)

	if err != nil {
		log.Errorf(c, "[budget.SavingsActualsHandler] failed to get transaction categories for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	categoryIds := make([]int64, 0, len(allCategories))

	for _, cat := range allCategories {
		if cat.ParentCategoryId != models.LevelOneTransactionCategoryParentId {
			categoryIds = append(categoryIds, cat.CategoryId)
		}
	}

	items, err := a.budgetTargets.GetSavingsActuals(c, uid, req.Year, req.Month, int(user.BudgetEndDay), categoryIds)

	if err != nil {
		log.Errorf(c, "[budget.SavingsActualsHandler] failed to get savings actuals for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return &models.SavingsActualsResponse{Items: items}, nil
}

// ExpenseIncomeActualsHandler returns budget-filtered expense and income actuals for the given time range
func (a *BudgetApi) ExpenseIncomeActualsHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.BudgetExpenseIncomeActualsGetRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Warnf(c, "[budget.ExpenseIncomeActualsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	items, err := a.budgetTargets.GetExpenseIncomeActuals(c, uid, req.StartTime, req.EndTime)

	if err != nil {
		log.Errorf(c, "[budget.ExpenseIncomeActualsHandler] failed to get expense/income actuals for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return &models.BudgetExpenseIncomeActualsResponse{Items: items}, nil
}

// BudgetTargetsHandler returns budget targets for the given year and month
func (a *BudgetApi) BudgetTargetsHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.BudgetTargetsGetRequest
	err := c.ShouldBindQuery(&req)

	if err != nil {
		log.Warnf(c, "[budget.BudgetTargetsHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	targets, err := a.budgetTargets.GetBudgetTargets(c, uid, req.Year, req.Month)

	if err != nil {
		log.Errorf(c, "[budget.BudgetTargetsHandler] failed to get budget targets for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if req.CategoryId > 0 {
		filtered := make([]*models.BudgetTarget, 0, len(targets))

		for i := 0; i < len(targets); i++ {
			if targets[i].CategoryId == req.CategoryId {
				filtered = append(filtered, targets[i])
			}
		}

		targets = filtered
	}

	resp := make([]*models.BudgetTargetInfoResponse, len(targets))

	for i := 0; i < len(targets); i++ {
		resp[i] = targets[i].ToInfoResponse()
	}

	return resp, nil
}

// CreateBudgetTargetHandler saves a new budget target by request parameters for current user
func (a *BudgetApi) CreateBudgetTargetHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.BudgetTargetCreateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[budget.CreateBudgetTargetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	target, err := a.budgetTargets.CreateBudgetTarget(c, uid, &req)

	if err != nil {
		log.Errorf(c, "[budget.CreateBudgetTargetHandler] failed to create budget target for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[budget.CreateBudgetTargetHandler] user \"uid:%d\" has created a new budget target \"id:%d\" successfully", uid, target.Id)

	return target.ToInfoResponse(), nil
}

// UpdateBudgetTargetHandler saves an existed budget target by request parameters for current user
func (a *BudgetApi) UpdateBudgetTargetHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.BudgetTargetUpdateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[budget.UpdateBudgetTargetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	target, err := a.budgetTargets.UpdateBudgetTarget(c, uid, &req)

	if err != nil {
		log.Errorf(c, "[budget.UpdateBudgetTargetHandler] failed to update budget target \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[budget.UpdateBudgetTargetHandler] user \"uid:%d\" has updated budget target \"id:%d\" successfully", uid, req.Id)

	return target.ToInfoResponse(), nil
}

// DeleteBudgetTargetHandler deletes an existed budget target by request parameters for current user
func (a *BudgetApi) DeleteBudgetTargetHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.BudgetTargetDeleteRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[budget.DeleteBudgetTargetHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.budgetTargets.DeleteBudgetTarget(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[budget.DeleteBudgetTargetHandler] failed to delete budget target \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[budget.DeleteBudgetTargetHandler] user \"uid:%d\" has deleted budget target \"id:%d\"", uid, req.Id)

	return true, nil
}

// SetTransactionBudgetOverrideHandler adds or removes the budget exclusion flag for a transaction
func (a *BudgetApi) SetTransactionBudgetOverrideHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.TransactionBudgetOverrideSetRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[budget.SetTransactionBudgetOverrideHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.transactionBudgetOverrides.SetExclusion(c, uid, req.TransactionId, req.Excluded)

	if err != nil {
		log.Errorf(c, "[budget.SetTransactionBudgetOverrideHandler] failed to set budget override for user \"uid:%d\" transaction \"id:%d\", because %s", uid, req.TransactionId, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[budget.SetTransactionBudgetOverrideHandler] user \"uid:%d\" set budget exclusion for transaction \"id:%d\" to %v", uid, req.TransactionId, req.Excluded)

	return true, nil
}
