package services

import (
	"xorm.io/xorm"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
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
