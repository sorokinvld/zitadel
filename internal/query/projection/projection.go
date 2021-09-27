package projection

import (
	"context"
	"database/sql"
	"time"

	"github.com/caos/zitadel/internal/eventstore"
	"github.com/caos/zitadel/internal/eventstore/handler"
	"github.com/caos/zitadel/internal/eventstore/handler/crdb"
	"github.com/caos/zitadel/internal/query/projection/org/owner"
)

const (
	currentSeqTable   = "projections.current_sequences"
	locksTable        = "projections.locks"
	failedEventsTable = "projections.failed_events"
)

func Start(ctx context.Context, sqlClient *sql.DB, es *eventstore.Eventstore, config Config) error {
	projectionConfig := crdb.StatementHandlerConfig{
		ProjectionHandlerConfig: handler.ProjectionHandlerConfig{
			HandlerConfig: handler.HandlerConfig{
				Eventstore: es,
			},
			RequeueEvery:     config.RequeueEvery.Duration,
			RetryFailedAfter: config.RetryFailedAfter.Duration,
		},
		Client:            sqlClient,
		SequenceTable:     currentSeqTable,
		LockTable:         locksTable,
		FailedEventsTable: failedEventsTable,
		MaxFailureCount:   config.MaxFailureCount,
		BulkLimit:         config.BulkLimit,
	}

	NewOrgProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["orgs"]))
	NewProjectProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["projects"]))
	owner.NewOrgOwnerProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["org_owners"]))
	NewPasswordComplexityProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["password_complexity_policy"]))
	NewPasswordAgeProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["password_age_policy"]))
	NewLockoutPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["lockout_policy"]))
	NewPrivacyPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["privacy_policy"]))
	NewOrgIAMPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["org_iam_policy"]))
	return nil
}

func applyCustomConfig(config crdb.StatementHandlerConfig, customConfig CustomConfig) crdb.StatementHandlerConfig {
	if customConfig.BulkLimit != nil {
		config.BulkLimit = *customConfig.BulkLimit
	}
	if customConfig.MaxFailureCount != nil {
		config.MaxFailureCount = *customConfig.MaxFailureCount
	}
	if customConfig.RequeueEvery != nil {
		config.RequeueEvery = customConfig.RequeueEvery.Duration
	}
	if customConfig.RetryFailedAfter != nil {
		config.RetryFailedAfter = customConfig.RetryFailedAfter.Duration
	}

	return config
}

func iteratorPool(workerCount int) chan func() {
	if workerCount <= 0 {
		return nil
	}

	queue := make(chan func())
	for i := 0; i < workerCount; i++ {
		go func() {
			for iteration := range queue {
				iteration()
				time.Sleep(2 * time.Second)
			}
		}()
	}
	return queue
}
