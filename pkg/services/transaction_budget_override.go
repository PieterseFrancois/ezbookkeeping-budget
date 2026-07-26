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

// TransactionBudgetOverrideService handles per-transaction budget override records
type TransactionBudgetOverrideService struct {
	ServiceUsingDB
	ServiceUsingUuid
}

// TransactionBudgetOverrides is the singleton service instance
var TransactionBudgetOverrides = &TransactionBudgetOverrideService{
	ServiceUsingDB: ServiceUsingDB{
		container: datastore.Container,
	},
	ServiceUsingUuid: ServiceUsingUuid{
		container: uuid.Container,
	},
}

// GetExcludedTransactionIds returns a map of transaction IDs (from the given slice) that are excluded from budget.
// Returns an empty map when transactionIds is empty.
func (s *TransactionBudgetOverrideService) GetExcludedTransactionIds(c core.Context, uid int64, transactionIds []int64) (map[int64]bool, error) {
	if len(transactionIds) == 0 {
		return map[int64]bool{}, nil
	}

	var overrides []*models.TransactionBudgetOverride
	err := s.UserDataDB(uid).NewSession(c).
		Select("transaction_id").
		Where("uid=? AND excluded=?", uid, true).
		In("transaction_id", transactionIds).
		Find(&overrides)

	if err != nil {
		return nil, err
	}

	result := make(map[int64]bool, len(overrides))
	for _, o := range overrides {
		result[o.TransactionId] = true
	}

	return result, nil
}

// IsExcluded returns whether a single transaction is excluded from budget calculations.
func (s *TransactionBudgetOverrideService) IsExcluded(c core.Context, uid int64, transactionId int64) (bool, error) {
	return s.UserDataDB(uid).NewSession(c).
		Where("uid=? AND transaction_id=? AND excluded=?", uid, transactionId, true).
		Exist(&models.TransactionBudgetOverride{})
}

// SetExclusion adds or removes the budget exclusion for a transaction.
// The operation is idempotent: setting the same state twice is a no-op.
func (s *TransactionBudgetOverrideService) SetExclusion(c core.Context, uid int64, transactionId int64, excluded bool) error {
	if uid <= 0 {
		return errs.ErrUserIdInvalid
	}

	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		var existing models.TransactionBudgetOverride
		has, err := sess.Where("uid=? AND transaction_id=?", uid, transactionId).Get(&existing)
		if err != nil {
			return err
		}

		now := time.Now().Unix()

		if has {
			if existing.Excluded == excluded {
				return nil // already in desired state
			}
			existing.Excluded = excluded
			existing.UpdatedUnixTime = now
			_, err = sess.ID(existing.OverrideId).Cols("excluded", "updated_unix_time").Update(&existing)
			return err
		}

		// No existing row — only insert if we're setting excluded=true
		// (no row is the canonical "not excluded" state, so inserting excluded=false is a no-op)
		if !excluded {
			return nil
		}

		override := &models.TransactionBudgetOverride{
			Uid:             uid,
			TransactionId:   transactionId,
			Excluded:        true,
			CreatedUnixTime: now,
			UpdatedUnixTime: now,
		}
		_, err = sess.Insert(override)
		return err
	})
}

// DeleteOverridesForTransaction removes all override records for a single transaction.
// Called when a transaction is deleted to prevent orphaned rows.
func (s *TransactionBudgetOverrideService) DeleteOverridesForTransaction(c core.Context, uid int64, transactionId int64) error {
	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Where("uid=? AND transaction_id=?", uid, transactionId).
			Delete(&models.TransactionBudgetOverride{})
		return err
	})
}

// DeleteAllOverridesForUser removes all override records for a user.
// Called when all user data is wiped.
func (s *TransactionBudgetOverrideService) DeleteAllOverridesForUser(c core.Context, uid int64) error {
	return s.UserDataDB(uid).DoTransaction(c, func(sess *xorm.Session) error {
		_, err := sess.Where("uid=?", uid).Delete(&models.TransactionBudgetOverride{})
		return err
	})
}
