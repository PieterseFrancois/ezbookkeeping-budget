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
	budgetTargets *services.BudgetService
}

// Initialize a budget api singleton instance
var (
	Budget = &BudgetApi{
		budgetTargets: services.BudgetTargets,
	}
)

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
