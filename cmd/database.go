package cmd

import (
	"github.com/urfave/cli/v3"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

// Database represents the database command
var Database = &cli.Command{
	Name:  "database",
	Usage: "ezBookkeeping database maintenance",
	Commands: []*cli.Command{
		{
			Name:   "update",
			Usage:  "Update database structure",
			Action: bindAction(updateDatabaseStructure),
		},
	},
}

func updateDatabaseStructure(c *core.CliContext) error {
	_, err := initializeSystem(c)

	if err != nil {
		return err
	}

	log.CliInfof(c, "[database.updateDatabaseStructure] starting maintaining")

	err = updateAllDatabaseTablesStructure(c)

	if err != nil {
		log.CliErrorf(c, "[database.updateDatabaseStructure] update database table structure failed, because %s", err.Error())
		return err
	}

	log.CliInfof(c, "[database.updateDatabaseStructure] all tables maintained successfully")
	return nil
}

func updateAllDatabaseTablesStructure(c *core.CliContext) error {
	var err error

	err = datastore.Container.UserStore.SyncStructs(new(models.User))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user table maintained successfully")

	err = datastore.Container.UserStore.SyncStructs(new(models.TwoFactor))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] two-factor table maintained successfully")

	err = datastore.Container.UserStore.SyncStructs(new(models.TwoFactorRecoveryCode))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] two-factor recovery code table maintained successfully")

	err = datastore.Container.TokenStore.SyncStructs(new(models.TokenRecord))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] token record table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Account))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] account table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Transaction))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionDraft))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction draft table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionCategory))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction category table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTagGroup))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag group table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTag))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTagIndex))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction tag index table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionTemplate))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction template table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionPictureInfo))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction picture table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserCustomIcon))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user custom icon table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserCustomExchangeRate))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user custom exchange rate table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserApplicationCloudSetting))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user application cloud settings table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.UserExternalAuth))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] user external auth table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.InsightsExplorer))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] insights explorer table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.BudgetTarget))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] budget target table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.TransactionBudgetOverride))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] transaction budget override table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Goal))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] goal table maintained successfully")

	err = datastore.Container.UserDataStore.SyncStructs(new(models.Subscription))

	if err != nil {
		return err
	}

	log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] subscription table maintained successfully")

	// Data migration: budget targets created before sections existed have an empty section.
	// This is guarded on section='' so it is idempotent and a no-op on every subsequent start.
	backfilled, err := services.BudgetTargets.BackfillBudgetTargetSections(c)

	if err != nil {
		return err
	}

	if backfilled > 0 {
		log.BootInfof(c, "[database.updateAllDatabaseTablesStructure] backfilled section on %d budget target(s)", backfilled)
	}

	return nil
}
