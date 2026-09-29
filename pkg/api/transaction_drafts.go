package api

import (
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// TransactionDraftsApi represents transaction draft api
type TransactionDraftsApi struct {
	ApiUsingConfig
	transactionDrafts          *services.TransactionDraftService
	transactions               *services.TransactionService
	transactionTags            *services.TransactionTagService
	transactionPictures        *services.TransactionPictureService
	transactionBudgetOverrides *services.TransactionBudgetOverrideService
}

// Initialize a transaction draft api singleton instance
var (
	TransactionDrafts = &TransactionDraftsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		transactionDrafts:          services.TransactionDrafts,
		transactions:               services.Transactions,
		transactionTags:            services.TransactionTags,
		transactionPictures:        services.TransactionPictures,
		transactionBudgetOverrides: services.TransactionBudgetOverrides,
	}
)

// attachDraftPictures fetches and attaches picture info for a draft response's picture ids, if any
func (a *TransactionDraftsApi) attachDraftPictures(c core.Context, uid int64, resp *models.TransactionDraftInfoResponse) {
	if resp == nil || len(resp.PictureIds) == 0 {
		return
	}

	pictureIds, err := utils.StringArrayToInt64Array(resp.PictureIds)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.attachDraftPictures] failed to parse picture ids for user \"uid:%d\", because %s", uid, err.Error())
		return
	}

	pictureInfos, err := a.transactionPictures.GetNewPictureInfosByPictureIds(c, uid, pictureIds)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.attachDraftPictures] failed to get picture infos for user \"uid:%d\", because %s", uid, err.Error())
		return
	}

	resp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
}

// TransactionDraftCreateHandler saves a new transaction draft by request parameters for current user
func (a *TransactionDraftsApi) TransactionDraftCreateHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDraftCreateReq models.TransactionDraftCreateRequest
	err := c.ShouldBindJSON(&transactionDraftCreateReq)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	tagIds, err := utils.StringArrayToInt64Array(transactionDraftCreateReq.TagIds)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] parse tag ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionTagIdInvalid
	}

	if len(tagIds) > models.MaximumTagsCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyTags
	}

	pictureIds, err := utils.StringArrayToInt64Array(transactionDraftCreateReq.PictureIds)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] parse picture ids failed, because %s", err.Error())
		return nil, errs.ErrTransactionPictureIdInvalid
	}

	if len(pictureIds) > models.MaximumPicturesCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyPictures
	}

	if transactionDraftCreateReq.Type < models.TRANSACTION_TYPE_MODIFY_BALANCE || transactionDraftCreateReq.Type > models.TRANSACTION_TYPE_TRANSFER {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] transaction type is invalid")
		return nil, errs.ErrTransactionTypeInvalid
	}

	if transactionDraftCreateReq.Type == models.TRANSACTION_TYPE_MODIFY_BALANCE && transactionDraftCreateReq.CategoryId != 0 {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] balance modification transaction cannot set category id")
		return nil, errs.ErrBalanceModificationTransactionCannotSetCategory
	}

	if transactionDraftCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionDraftCreateReq.DestinationAccountId != 0 {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] non-transfer transaction destination account cannot be set")
		return nil, errs.ErrTransactionDestinationAccountCannotBeSet
	} else if transactionDraftCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER && transactionDraftCreateReq.AccountId != 0 && transactionDraftCreateReq.AccountId == transactionDraftCreateReq.DestinationAccountId {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] transfer transaction source account must not be destination account")
		return nil, errs.ErrTransactionSourceAndDestinationIdCannotBeEqual
	}

	if transactionDraftCreateReq.Type != models.TRANSACTION_TYPE_TRANSFER && transactionDraftCreateReq.DestinationAmount != 0 {
		log.Warnf(c, "[transaction_drafts.TransactionDraftCreateHandler] non-transfer transaction destination amount cannot be set")
		return nil, errs.ErrTransactionDestinationAmountCannotBeSet
	}

	transactionDbType, err := transactionDraftCreateReq.Type.ToTransactionDbType()

	if err != nil {
		return nil, errs.ErrTransactionTypeInvalid
	}

	uid := c.GetCurrentUid()

	draft := &models.TransactionDraft{
		Uid:               uid,
		Type:              transactionDbType,
		CategoryId:        transactionDraftCreateReq.CategoryId,
		AccountId:         transactionDraftCreateReq.AccountId,
		TransactionTime:   utils.GetMinTransactionTimeFromUnixTime(transactionDraftCreateReq.Time),
		TimezoneUtcOffset: transactionDraftCreateReq.UtcOffset,
		Amount:            transactionDraftCreateReq.Amount,
		HideAmount:        transactionDraftCreateReq.HideAmount,
		TagIds:            strings.Join(utils.Int64ArrayToStringArray(utils.ToUniqueInt64Slice(tagIds)), ","),
		PictureIds:        strings.Join(utils.Int64ArrayToStringArray(utils.ToUniqueInt64Slice(pictureIds)), ","),
		Comment:           transactionDraftCreateReq.Comment,
		ExcludeFromBudget: transactionDraftCreateReq.ExcludeFromBudget,
		Source:            transactionDraftCreateReq.Source,
		CreatedIp:         c.ClientIP(),
	}

	if transactionDraftCreateReq.Type == models.TRANSACTION_TYPE_TRANSFER {
		draft.DestinationAccountId = transactionDraftCreateReq.DestinationAccountId
		draft.DestinationAmount = transactionDraftCreateReq.DestinationAmount
	}

	err = a.transactionDrafts.CreateTransactionDraft(c, draft)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transaction_drafts.TransactionDraftCreateHandler] failed to create transaction draft for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_drafts.TransactionDraftCreateHandler] user \"uid:%d\" has created a new transaction draft \"id:%d\" successfully", uid, draft.TransactionDraftId)

	draftResp := draft.ToTransactionDraftInfoResponse()
	a.attachDraftPictures(c, uid, draftResp)

	return draftResp, nil
}

// TransactionDraftGetBySourceHandler returns a transaction draft or confirmed transaction, looked up by its source, for current user
func (a *TransactionDraftsApi) TransactionDraftGetBySourceHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDraftGetReq models.TransactionDraftGetRequest
	err := c.ShouldBindJSON(&transactionDraftGetReq)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftGetBySourceHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	resp, err := a.transactionDrafts.GetTransactionOrDraftBySource(c, uid, transactionDraftGetReq.Source)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transaction_drafts.TransactionDraftGetBySourceHandler] failed to get transaction or draft for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if resp.Status == "draft" && resp.Draft != nil {
		a.attachDraftPictures(c, uid, resp.Draft)
	}

	return resp, nil
}

// TransactionDraftModifyHandler modifies an existed transaction draft (or, if already confirmed, the
// underlying confirmed transaction), looked up by its source, by request parameters for current user
func (a *TransactionDraftsApi) TransactionDraftModifyHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDraftModifyReq models.TransactionDraftModifyRequest
	err := c.ShouldBindJSON(&transactionDraftModifyReq)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftModifyHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if transactionDraftModifyReq.TagIds != nil && len(transactionDraftModifyReq.TagIds) > models.MaximumTagsCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyTags
	}

	if transactionDraftModifyReq.PictureIds != nil && len(transactionDraftModifyReq.PictureIds) > models.MaximumPicturesCountOfTransaction {
		return nil, errs.ErrTransactionHasTooManyPictures
	}

	uid := c.GetCurrentUid()

	draft, err := a.transactionDrafts.ModifyTransactionDraftBySource(c, uid, transactionDraftModifyReq.Source, &transactionDraftModifyReq)

	if err == nil {
		log.Infof(c, "[transaction_drafts.TransactionDraftModifyHandler] user \"uid:%d\" has modified transaction draft \"source:%s\" successfully", uid, transactionDraftModifyReq.Source)
		draftResp := draft.ToTransactionDraftInfoResponse()
		a.attachDraftPictures(c, uid, draftResp)
		return draftResp, nil
	}

	if err != errs.ErrTransactionDraftNotFound {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transaction_drafts.TransactionDraftModifyHandler] failed to modify transaction draft for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	// No draft found for this source - fall back to modifying the underlying confirmed transaction, if any.
	transactionResp, modifyErr := a.modifyConfirmedTransactionBySource(c, uid, &transactionDraftModifyReq)

	if modifyErr != nil {
		return nil, modifyErr
	}

	return transactionResp, nil
}

// modifyConfirmedTransactionBySource applies a draft-modify-shaped partial update onto an already
// confirmed transaction (looked up by source), delegating the actual persistence to the existing,
// unmodified TransactionService.ModifyTransaction.
func (a *TransactionDraftsApi) modifyConfirmedTransactionBySource(c *core.WebContext, uid int64, req *models.TransactionDraftModifyRequest) (*models.TransactionInfoResponse, *errs.Error) {
	oldTransaction, err := a.transactions.GetTransactionBySource(c, uid, req.Source)

	if err != nil {
		if err == errs.ErrTransactionNotFound {
			return nil, errs.ErrTransactionDraftNotFound
		}

		log.Errorf(c, "[transaction_drafts.modifyConfirmedTransactionBySource] failed to get transaction by source for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if oldTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transaction_drafts.modifyConfirmedTransactionBySource] cannot modify transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", oldTransaction.TransactionId, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	allTransactionTagIds, err := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{oldTransaction.TransactionId})

	if err != nil {
		log.Errorf(c, "[transaction_drafts.modifyConfirmedTransactionBySource] failed to get transaction tag ids for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	oldTagIds := allTransactionTagIds[oldTransaction.TransactionId]

	if oldTagIds == nil {
		oldTagIds = make([]int64, 0)
	}

	oldPictureInfos, err := a.transactionPictures.GetPictureInfosByTransactionId(c, uid, oldTransaction.TransactionId)

	if err != nil {
		log.Errorf(c, "[transaction_drafts.modifyConfirmedTransactionBySource] failed to get transaction picture infos for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	oldPictureIds := a.transactionPictures.GetTransactionPictureIds(oldPictureInfos)

	newTransaction := &models.Transaction{
		TransactionId:        oldTransaction.TransactionId,
		Uid:                  uid,
		Type:                 oldTransaction.Type,
		CategoryId:           oldTransaction.CategoryId,
		TransactionTime:      oldTransaction.TransactionTime,
		TimezoneUtcOffset:    oldTransaction.TimezoneUtcOffset,
		AccountId:            oldTransaction.AccountId,
		Amount:               oldTransaction.Amount,
		RelatedAccountId:     oldTransaction.RelatedAccountId,
		RelatedAccountAmount: oldTransaction.RelatedAccountAmount,
		HideAmount:           oldTransaction.HideAmount,
		Comment:              oldTransaction.Comment,
	}

	if req.Type != nil {
		newType, typeErr := (*req.Type).ToTransactionDbType()

		if typeErr != nil {
			return nil, errs.ErrTransactionTypeInvalid
		}

		newTransaction.Type = newType
	}

	if req.CategoryId != nil {
		newTransaction.CategoryId = *req.CategoryId
	}

	if req.Time != nil {
		newTransaction.TransactionTime = utils.GetMinTransactionTimeFromUnixTime(*req.Time)
	}

	if req.UtcOffset != nil {
		newTransaction.TimezoneUtcOffset = *req.UtcOffset
	}

	if req.AccountId != nil {
		newTransaction.AccountId = *req.AccountId
	}

	if req.Amount != nil {
		newTransaction.Amount = *req.Amount
	}

	if req.DestinationAccountId != nil {
		newTransaction.RelatedAccountId = *req.DestinationAccountId
	}

	if req.DestinationAmount != nil {
		newTransaction.RelatedAccountAmount = *req.DestinationAmount
	}

	if req.HideAmount != nil {
		newTransaction.HideAmount = *req.HideAmount
	}

	if req.Comment != nil {
		newTransaction.Comment = *req.Comment
	}

	newTagIds := oldTagIds
	var err2 error

	if req.TagIds != nil {
		newTagIds, err2 = utils.StringArrayToInt64Array(req.TagIds)

		if err2 != nil {
			return nil, errs.ErrTransactionTagIdInvalid
		}
	}

	newPictureIds := oldPictureIds

	if req.PictureIds != nil {
		newPictureIds, err2 = utils.StringArrayToInt64Array(req.PictureIds)

		if err2 != nil {
			return nil, errs.ErrTransactionPictureIdInvalid
		}
	}

	changeToTransfer := newTransaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_OUT && oldTransaction.Type != models.TRANSACTION_DB_TYPE_TRANSFER_OUT

	var addTagIds, removeTagIds []int64

	if !utils.Int64SliceEquals(newTagIds, oldTagIds) {
		removeTagIds = oldTagIds
		addTagIds = newTagIds
	}

	addPictureIds := utils.Int64SliceMinus(newPictureIds, oldPictureIds)
	removePictureIds := utils.Int64SliceMinus(oldPictureIds, newPictureIds)

	err = a.transactions.ModifyTransaction(c, newTransaction, changeToTransfer, len(oldTagIds), addTagIds, removeTagIds, addPictureIds, removePictureIds)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transaction_drafts.modifyConfirmedTransactionBySource] failed to modify transaction \"id:%d\" for user \"uid:%d\", because %s", oldTransaction.TransactionId, uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_drafts.modifyConfirmedTransactionBySource] user \"uid:%d\" has modified confirmed transaction \"id:%d\" (source:%s) successfully", uid, oldTransaction.TransactionId, req.Source)

	return newTransaction.ToTransactionInfoResponse(newTagIds, true), nil
}

// TransactionDraftDeleteHandler deletes an existed transaction draft (or, if already confirmed, the
// underlying confirmed transaction), looked up by its source, for current user
func (a *TransactionDraftsApi) TransactionDraftDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDraftDeleteReq models.TransactionDraftDeleteRequest
	err := c.ShouldBindJSON(&transactionDraftDeleteReq)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftDeleteHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()

	_, err = a.transactionDrafts.GetTransactionDraftBySource(c, uid, transactionDraftDeleteReq.Source)

	if err == nil {
		err = a.transactionDrafts.DeleteTransactionDraftBySource(c, uid, transactionDraftDeleteReq.Source)

		if err != nil {
			log.Errorf(c, "[transaction_drafts.TransactionDraftDeleteHandler] failed to delete transaction draft for user \"uid:%d\", because %s", uid, err.Error())
			return nil, errs.Or(err, errs.ErrOperationFailed)
		}

		log.Infof(c, "[transaction_drafts.TransactionDraftDeleteHandler] user \"uid:%d\" has deleted transaction draft \"source:%s\" successfully", uid, transactionDraftDeleteReq.Source)
		return true, nil
	}

	if err != errs.ErrTransactionDraftNotFound {
		log.Errorf(c, "[transaction_drafts.TransactionDraftDeleteHandler] failed to get transaction draft for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	// No draft found for this source - fall back to deleting the underlying confirmed transaction, if any.
	transaction, err := a.transactions.GetTransactionBySource(c, uid, transactionDraftDeleteReq.Source)

	if err != nil {
		if err == errs.ErrTransactionNotFound {
			return nil, errs.ErrTransactionDraftNotFound
		}

		log.Errorf(c, "[transaction_drafts.TransactionDraftDeleteHandler] failed to get transaction by source for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if transaction.Type == models.TRANSACTION_DB_TYPE_TRANSFER_IN {
		log.Warnf(c, "[transaction_drafts.TransactionDraftDeleteHandler] cannot delete transaction \"id:%d\" for user \"uid:%d\", because transaction type is transfer in", transaction.TransactionId, uid)
		return nil, errs.ErrTransactionTypeInvalid
	}

	err = a.transactions.DeleteTransaction(c, uid, transaction.TransactionId)

	if err != nil {
		log.Errorf(c, "[transaction_drafts.TransactionDraftDeleteHandler] failed to delete transaction \"id:%d\" for user \"uid:%d\", because %s", transaction.TransactionId, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_drafts.TransactionDraftDeleteHandler] user \"uid:%d\" has deleted confirmed transaction \"id:%d\" (source:%s) successfully", uid, transaction.TransactionId, transactionDraftDeleteReq.Source)
	return true, nil
}

// TransactionDraftListHandler returns paginated transaction drafts of current user
func (a *TransactionDraftsApi) TransactionDraftListHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDraftListReq models.TransactionDraftListRequest
	err := c.ShouldBindQuery(&transactionDraftListReq)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftListHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	drafts, err := a.transactionDrafts.ListTransactionDraftsByUser(c, uid, transactionDraftListReq.Page, transactionDraftListReq.Count)

	if err != nil {
		log.Errorf(c, "[transaction_drafts.TransactionDraftListHandler] failed to list transaction drafts for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	draftResps := make([]*models.TransactionDraftInfoResponse, len(drafts))

	for i := 0; i < len(drafts); i++ {
		draftResps[i] = drafts[i].ToTransactionDraftInfoResponse()
	}

	a.attachDraftPicturesBatch(c, uid, draftResps)

	return draftResps, nil
}

// attachDraftPicturesBatch fetches and attaches picture info for a batch of draft responses in a
// single query, then distributes results back to each response by its own picture ids
func (a *TransactionDraftsApi) attachDraftPicturesBatch(c core.Context, uid int64, draftResps []*models.TransactionDraftInfoResponse) {
	allPictureIdsSet := make(map[string]bool)

	for _, draftResp := range draftResps {
		for _, pictureId := range draftResp.PictureIds {
			allPictureIdsSet[pictureId] = true
		}
	}

	if len(allPictureIdsSet) == 0 {
		return
	}

	allPictureIdStrs := make([]string, 0, len(allPictureIdsSet))

	for pictureId := range allPictureIdsSet {
		allPictureIdStrs = append(allPictureIdStrs, pictureId)
	}

	allPictureIds, err := utils.StringArrayToInt64Array(allPictureIdStrs)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.attachDraftPicturesBatch] failed to parse picture ids for user \"uid:%d\", because %s", uid, err.Error())
		return
	}

	pictureInfos, err := a.transactionPictures.GetNewPictureInfosByPictureIds(c, uid, allPictureIds)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.attachDraftPicturesBatch] failed to get picture infos for user \"uid:%d\", because %s", uid, err.Error())
		return
	}

	pictureInfoMap := a.transactionPictures.GetPictureInfoMapByList(pictureInfos)

	for _, draftResp := range draftResps {
		if len(draftResp.PictureIds) == 0 {
			continue
		}

		matchedPictureInfos := make([]*models.TransactionPictureInfo, 0, len(draftResp.PictureIds))

		for _, pictureIdStr := range draftResp.PictureIds {
			pictureId, parseErr := utils.StringToInt64(pictureIdStr)

			if parseErr != nil {
				continue
			}

			if pictureInfo, exists := pictureInfoMap[pictureId]; exists {
				matchedPictureInfos = append(matchedPictureInfos, pictureInfo)
			}
		}

		draftResp.Pictures = a.GetTransactionPictureInfoResponseList(matchedPictureInfos)
	}
}

// TransactionDraftCountHandler returns the total number of pending transaction drafts for current user
func (a *TransactionDraftsApi) TransactionDraftCountHandler(c *core.WebContext) (any, *errs.Error) {
	uid := c.GetCurrentUid()
	count, err := a.transactionDrafts.GetTransactionDraftCount(c, uid)

	if err != nil {
		log.Errorf(c, "[transaction_drafts.TransactionDraftCountHandler] failed to get transaction draft count for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return &models.TransactionDraftCountResponse{TotalCount: count}, nil
}

// TransactionDraftConfirmHandler promotes an existed transaction draft into a real, confirmed transaction for current user
func (a *TransactionDraftsApi) TransactionDraftConfirmHandler(c *core.WebContext) (any, *errs.Error) {
	var transactionDraftConfirmReq models.TransactionDraftConfirmRequest
	err := c.ShouldBindJSON(&transactionDraftConfirmReq)

	if err != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftConfirmHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	uid := c.GetCurrentUid()
	transaction, err := a.transactionDrafts.ConfirmTransactionDraft(c, uid, transactionDraftConfirmReq.Source, &transactionDraftConfirmReq)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Errorf(c, "[transaction_drafts.TransactionDraftConfirmHandler] failed to confirm transaction draft for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	log.Infof(c, "[transaction_drafts.TransactionDraftConfirmHandler] user \"uid:%d\" has confirmed transaction draft \"source:%s\" as transaction \"id:%d\" successfully", uid, transactionDraftConfirmReq.Source, transaction.TransactionId)

	var tagIds []int64
	allTransactionTagIds, tagErr := a.transactionTags.GetAllTagIdsOfTransactions(c, uid, []int64{transaction.TransactionId})

	if tagErr == nil {
		tagIds = allTransactionTagIds[transaction.TransactionId]
	}

	transactionResp := transaction.ToTransactionInfoResponse(tagIds, true)

	excluded, excludedErr := a.transactionBudgetOverrides.IsExcluded(c, uid, transaction.TransactionId)

	if excludedErr != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftConfirmHandler] failed to get budget override for user \"uid:%d\", because %s", uid, excludedErr.Error())
	} else {
		transactionResp.ExcludeFromBudget = excluded
	}

	pictureInfos, pictureErr := a.transactionPictures.GetPictureInfosByTransactionId(c, uid, transaction.TransactionId)

	if pictureErr != nil {
		log.Warnf(c, "[transaction_drafts.TransactionDraftConfirmHandler] failed to get pictures for user \"uid:%d\", because %s", uid, pictureErr.Error())
	} else {
		transactionResp.Pictures = a.GetTransactionPictureInfoResponseList(pictureInfos)
	}

	return transactionResp, nil
}
