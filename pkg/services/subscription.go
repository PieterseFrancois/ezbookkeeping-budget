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

// SubscriptionService represents subscription service
type SubscriptionService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// Initialize a subscription service singleton instance
var (
	Subscriptions = &SubscriptionService{
		ServiceUsingDB: ServiceUsingDB{
			container: datastore.Container,
		},
		ServiceUsingUuid: ServiceUsingUuid{
			container: uuid.Container,
		},
	}
)

// GetSubscriptions returns all subscriptions for the given user
func (s *SubscriptionService) GetSubscriptions(c core.Context, uid int64) ([]*models.Subscription, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	var subscriptions []*models.Subscription
	err := s.UserDataDB(uid).NewSession(c).Where("uid=?", uid).Find(&subscriptions)

	return subscriptions, err
}

// CreateSubscription saves a new subscription model to database
func (s *SubscriptionService) CreateSubscription(c core.Context, uid int64, request *models.SubscriptionCreateRequest) (*models.Subscription, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	subscription := &models.Subscription{
		Id:         s.GenerateUuid(uuid.UUID_TYPE_SUBSCRIPTION),
		Uid:        uid,
		Name:       request.Name,
		Amount:     request.Amount,
		Currency:   request.Currency,
		TemplateId: request.TemplateId,
		StartDate:  request.StartDate,
		Frequency:  request.Frequency,
		IsActive:   request.IsActive,
		CreatedAt:  time.Now().Unix(),
	}

	if subscription.Id < 1 {
		return nil, errs.ErrSystemIsBusy
	}

	err := s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		err := s.isSubscriptionValid(sess, subscription)

		if err != nil {
			return err
		}

		_, err = sess.Insert(subscription)
		return err
	})

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// UpdateSubscription saves an existed subscription model to database
func (s *SubscriptionService) UpdateSubscription(c core.Context, uid int64, request *models.SubscriptionUpdateRequest) (*models.Subscription, error) {
	if uid <= 0 {
		return nil, errs.ErrUserIdInvalid
	}

	subscription := &models.Subscription{}
	has, err := s.UserDataDB(uid).NewSession(c).ID(request.Id).Where("uid=?", uid).Get(subscription)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errs.ErrSubscriptionNotFound
	}

	subscription.Name = request.Name
	subscription.Amount = request.Amount
	subscription.Currency = request.Currency
	subscription.TemplateId = request.TemplateId
	subscription.StartDate = request.StartDate
	subscription.Frequency = request.Frequency
	subscription.IsActive = request.IsActive

	err = s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		err := s.isSubscriptionValid(sess, subscription)

		if err != nil {
			return err
		}

		updatedRows, err := sess.ID(subscription.Id).Cols("name", "amount", "currency", "template_id", "start_date", "frequency", "is_active").Where("uid=?", uid).Update(subscription)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrSubscriptionNotFound
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// ToggleSubscriptionActive updates the active state of an existed subscription
func (s *SubscriptionService) ToggleSubscriptionActive(c core.Context, uid int64, id int64, isActive bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return errs.ErrSubscriptionIdInvalid
	}

	updateModel := &models.Subscription{
		IsActive: isActive,
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		updatedRows, err := sess.ID(id).Cols("is_active").Where("uid=?", uid).Update(updateModel)

		if err != nil {
			return err
		} else if updatedRows < 1 {
			return errs.ErrSubscriptionNotFound
		}

		return nil
	})
}

// DeleteSubscription deletes an existed subscription from database
func (s *SubscriptionService) DeleteSubscription(c core.Context, uid int64, id int64) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	if id <= 0 {
		return errs.ErrSubscriptionIdInvalid
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		deletedRows, err := sess.ID(id).Where("uid=?", uid).Delete(&models.Subscription{})

		if err != nil {
			return err
		} else if deletedRows < 1 {
			return errs.ErrSubscriptionNotFound
		}

		return nil
	})
}

func (s *SubscriptionService) isSubscriptionValid(sess *xorm.Session, subscription *models.Subscription) error {
	if subscription.TemplateId != 0 {
		template := &models.TransactionTemplate{}
		has, err := sess.ID(subscription.TemplateId).Where("uid=? AND deleted=?", subscription.Uid, false).Get(template)

		if err != nil {
			return err
		} else if !has || template.TemplateType != models.TRANSACTION_TEMPLATE_TYPE_NORMAL {
			return errs.ErrSubscriptionTemplateInvalid
		}
	}

	return nil
}
