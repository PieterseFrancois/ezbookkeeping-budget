package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// GoalApi represents goal api
type GoalApi struct {
	goals *services.GoalService
}

// Initialize a goal api singleton instance
var (
	Goal = &GoalApi{
		goals: services.Goals,
	}
)

// GoalsHandler returns all goals for current user
func (a *GoalApi) GoalsHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	goals, err := a.goals.GetGoals(c, uid)

	if err != nil {
		log.Errorf(c, "[goal.GoalsHandler] failed to get goals for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resp := make([]*models.GoalInfoResponse, len(goals))

	for i := 0; i < len(goals); i++ {
		resp[i] = goals[i].ToInfoResponse()
	}

	return resp, nil
}

// CreateGoalHandler saves a new goal by request parameters for current user
func (a *GoalApi) CreateGoalHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.GoalCreateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[goal.CreateGoalHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	goal, err := a.goals.CreateGoal(c, uid, &req)

	if err != nil {
		log.Errorf(c, "[goal.CreateGoalHandler] failed to create goal for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[goal.CreateGoalHandler] user \"uid:%d\" has created a new goal \"id:%d\" successfully", uid, goal.Id)

	return goal.ToInfoResponse(), nil
}

// UpdateGoalHandler saves an existed goal by request parameters for current user
func (a *GoalApi) UpdateGoalHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.GoalUpdateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[goal.UpdateGoalHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	goal, err := a.goals.UpdateGoal(c, uid, &req)

	if err != nil {
		log.Errorf(c, "[goal.UpdateGoalHandler] failed to update goal \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[goal.UpdateGoalHandler] user \"uid:%d\" has updated goal \"id:%d\" successfully", uid, req.Id)

	return goal.ToInfoResponse(), nil
}

// DeleteGoalHandler deletes an existed goal by request parameters for current user
func (a *GoalApi) DeleteGoalHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.GoalDeleteRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[goal.DeleteGoalHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.goals.DeleteGoal(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[goal.DeleteGoalHandler] failed to delete goal \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[goal.DeleteGoalHandler] user \"uid:%d\" has deleted goal \"id:%d\"", uid, req.Id)

	return true, nil
}
