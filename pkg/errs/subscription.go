package errs

import "net/http"

// Error codes related to subscriptions
var (
	ErrSubscriptionIdInvalid       = NewNormalError(NormalSubcategorySubscription, 0, http.StatusBadRequest, "subscription id is invalid")
	ErrSubscriptionNotFound        = NewNormalError(NormalSubcategorySubscription, 1, http.StatusBadRequest, "subscription not found")
	ErrSubscriptionCategoryInvalid = NewNormalError(NormalSubcategorySubscription, 2, http.StatusBadRequest, "subscription category is invalid")
	ErrSubscriptionTemplateInvalid = NewNormalError(NormalSubcategorySubscription, 3, http.StatusBadRequest, "subscription template is invalid")
)
