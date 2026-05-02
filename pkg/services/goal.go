package services

import (
	"time"

	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/uuid"
)

// GoalService represents goal service
type GoalService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a goal service singleton instance
var (
	Goals = &GoalService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetGoals returns all goals for the given user
func (s *GoalService) GetGoals(c core.Context, uid int64) ([]*models.Goal, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var goals []*models.Goal
	err := s.UserDataDB(uid).NewSession(c).Where("uid=?", uid).Find(&goals)

	return goals, err
}

// CreateGoal saves a new goal model to database
func (s *GoalService) CreateGoal(c core.Context, uid int64, request *models.GoalCreateRequest) (*models.Goal, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	goal := &models.Goal{
		Id:           s.GenerateUuid(uuid.UUID_TYPE_GOAL),
		Uid:          uid,
		Name:         request.Name,
		AccountId:    request.AccountId,
		TargetAmount: request.TargetAmount,
		TargetDate:   request.TargetDate,
		CreatedAt:    time.Now().Unix(),
	}

	if goal.Id < 1 {
		return nil, errs.ErrSystemIsBusy
	}

	err := s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Insert(goal)
		return err
	})

	if err != nil {
		return nil, err
	}

	return goal, nil
}

// UpdateGoal saves an existed goal model to database
func (s *GoalService) UpdateGoal(c core.Context, uid int64, request *models.GoalUpdateRequest) (*models.Goal, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	goal := &models.Goal{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(request.Id).Where("uid=?", uid).Get(goal)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrGoalNotFound
	}

	goal.Name = request.Name
	goal.AccountId = request.AccountId
	goal.TargetAmount = request.TargetAmount
	goal.TargetDate = request.TargetDate

	err = s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(goal.Id).Cols("name", "account_id", "target_amount", "target_date").Where("uid=?", uid).Update(goal)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrGoalNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return goal, nil
}

// DeleteGoal deletes an existed goal from database
func (s *GoalService) DeleteGoal(c core.Context, uid int64, id int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return errs.ErrGoalIdInvalid
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(id).Where("uid=?", uid).Delete(&models.Goal{})

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrGoalNotFound
		}

		return nil
	})
}
