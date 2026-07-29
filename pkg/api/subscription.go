package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// SubscriptionApi represents subscription api
type SubscriptionApi struct {
	subscriptions *services.SubscriptionService
}

// Initialize a subscription api singleton instance
var (
	Subscription = &SubscriptionApi{
		subscriptions: services.Subscriptions,
	}
)

// SubscriptionsHandler returns all subscriptions for current user
func (a *SubscriptionApi) SubscriptionsHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	subscriptions, err := a.subscriptions.GetSubscriptions(c, uid)

	if err != nil {
		log.Errorf(c, "[subscription.SubscriptionsHandler] failed to get subscriptions for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	resp := make([]*models.SubscriptionInfoResponse, len(subscriptions))

	for i := 0; i < len(subscriptions); i++ {
		resp[i] = subscriptions[i].ToInfoResponse()
	}

	return resp, nil
}

// CreateSubscriptionHandler saves a new subscription by request parameters for current user
func (a *SubscriptionApi) CreateSubscriptionHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.SubscriptionCreateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[subscription.CreateSubscriptionHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	subscription, err := a.subscriptions.CreateSubscription(c, uid, &req)

	if err != nil {
		log.Errorf(c, "[subscription.CreateSubscriptionHandler] failed to create subscription for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[subscription.CreateSubscriptionHandler] user \"uid:%d\" has created a new subscription \"id:%d\" successfully", uid, subscription.Id)

	return subscription.ToInfoResponse(), nil
}

// UpdateSubscriptionHandler saves an existed subscription by request parameters for current user
func (a *SubscriptionApi) UpdateSubscriptionHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.SubscriptionUpdateRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[subscription.UpdateSubscriptionHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	subscription, err := a.subscriptions.UpdateSubscription(c, uid, &req)

	if err != nil {
		log.Errorf(c, "[subscription.UpdateSubscriptionHandler] failed to update subscription \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[subscription.UpdateSubscriptionHandler] user \"uid:%d\" has updated subscription \"id:%d\" successfully", uid, req.Id)

	return subscription.ToInfoResponse(), nil
}

// ToggleSubscriptionActiveHandler toggles the active state of an existed subscription by request parameters for current user
func (a *SubscriptionApi) ToggleSubscriptionActiveHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.SubscriptionToggleActiveRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[subscription.ToggleSubscriptionActiveHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.subscriptions.ToggleSubscriptionActive(c, uid, req.Id, req.IsActive)

	if err != nil {
		log.Errorf(c, "[subscription.ToggleSubscriptionActiveHandler] failed to toggle subscription \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[subscription.ToggleSubscriptionActiveHandler] user \"uid:%d\" has toggled subscription \"id:%d\" successfully", uid, req.Id)

	return true, nil
}

// DeleteSubscriptionHandler deletes an existed subscription by request parameters for current user
func (a *SubscriptionApi) DeleteSubscriptionHandler(c *core.WebContext) (any, *errs.Error) {
	var req models.SubscriptionDeleteRequest
	err := c.ShouldBindJSON(&req)

	if err != nil {
		log.Warnf(c, "[subscription.DeleteSubscriptionHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	err = a.subscriptions.DeleteSubscription(c, uid, req.Id)

	if err != nil {
		log.Errorf(c, "[subscription.DeleteSubscriptionHandler] failed to delete subscription \"id:%d\" for user \"uid:%d\", because %s", req.Id, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[subscription.DeleteSubscriptionHandler] user \"uid:%d\" has deleted subscription \"id:%d\"", uid, req.Id)

	return true, nil
}
