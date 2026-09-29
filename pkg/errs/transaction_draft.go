package errs

import "net/http"

// Error codes related to transaction drafts
var (
	ErrTransactionDraftSourceAlreadyExists = NewNormalError(NormalSubcategoryTransactionDraft, 0, http.StatusBadRequest, "transaction draft source already exists")
	ErrTransactionDraftSourceIsEmpty       = NewNormalError(NormalSubcategoryTransactionDraft, 1, http.StatusBadRequest, "transaction draft source is empty")
	ErrTransactionDraftNotFound            = NewNormalError(NormalSubcategoryTransactionDraft, 2, http.StatusBadRequest, "transaction draft not found")
	ErrTransactionDraftIncomplete          = NewNormalError(NormalSubcategoryTransactionDraft, 3, http.StatusBadRequest, "transaction draft is incomplete, account id and/or category id is not set")
)
