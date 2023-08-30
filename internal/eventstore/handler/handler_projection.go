package handler

import (
	"context"
	"errors"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zitadel/logging"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/api/call"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/repository/pseudo"
)

const (
	schedulerSucceeded = eventstore.EventType("system.projections.scheduler.succeeded")
	aggregateType      = eventstore.AggregateType("system")
	aggregateID        = "SYSTEM"
)

type ProjectionHandlerConfig struct {
	HandlerConfig
	ProjectionName        string
	RequeueEvery          time.Duration
	RetryFailedAfter      time.Duration
	Retries               uint
	ConcurrentInstances   uint
	HandleActiveInstances time.Duration
}

// Update updates the projection with the given statements
type Update func(context.Context, []*Statement, Reduce) (index int, err error)

// Reduce reduces the given event to a statement
// which is used to update the projection
type Reduce func(eventstore.Event) (*Statement, error)

// SearchQuery generates the search query to lookup for events
type SearchQuery func(ctx context.Context, instanceIDs []string) (query *eventstore.SearchQueryBuilder, queryLimit uint64, err error)

// Lock is used for mutex handling if needed on the projection
type Lock func(context.Context, time.Duration, ...string) <-chan error

// Unlock releases the mutex of the projection
type Unlock func(...string) error

// NowFunc makes time.Now() mockable
type NowFunc func() time.Time

type ProjectionHandler struct {
	Handler
	ProjectionName             string
	reduce                     Reduce
	update                     Update
	searchQuery                SearchQuery
	triggerProjection          *time.Timer
	lock                       Lock
	unlock                     Unlock
	requeueAfter               time.Duration
	retryFailedAfter           time.Duration
	retries                    int
	concurrentInstances        int
	handleActiveInstances      time.Duration
	nowFunc                    NowFunc
	reduceScheduledPseudoEvent bool
}

func NewProjectionHandler(
	ctx context.Context,
	config ProjectionHandlerConfig,
	reduce Reduce,
	update Update,
	query SearchQuery,
	lock Lock,
	unlock Unlock,
	initialized <-chan bool,
	reduceScheduledPseudoEvent bool,
) *ProjectionHandler {
	concurrentInstances := int(config.ConcurrentInstances)
	if concurrentInstances < 1 {
		concurrentInstances = 1
	}
	h := &ProjectionHandler{
		Handler:                    NewHandler(config.HandlerConfig),
		ProjectionName:             config.ProjectionName,
		reduce:                     reduce,
		update:                     update,
		searchQuery:                query,
		lock:                       lock,
		unlock:                     unlock,
		requeueAfter:               config.RequeueEvery,
		triggerProjection:          time.NewTimer(0), // first trigger is instant on startup
		retryFailedAfter:           config.RetryFailedAfter,
		retries:                    int(config.Retries),
		concurrentInstances:        concurrentInstances,
		handleActiveInstances:      config.HandleActiveInstances,
		nowFunc:                    time.Now,
		reduceScheduledPseudoEvent: reduceScheduledPseudoEvent,
	}

	go func() {
		<-initialized
		if !h.reduceScheduledPseudoEvent {
			go h.subscribe(ctx)
		}
		go h.schedule(ctx)
	}()

	return h
}

func triggerInstances(ctx context.Context, instances []string) []string {
	if len(instances) == 0 {
		instances = append(instances, authz.GetInstance(ctx).InstanceID())
	}
	return instances
}

// Trigger handles all events for the provided instances (or current instance from context if non specified)
// by calling FetchEvents and Process until the amount of events is smaller than the BulkLimit.
// If a bulk action was executed, the call timestamp in context will be reset for subsequent queries.
// The returned context is never nil. It is either the original context or an updated context.
//
// If Trigger encounters an error, it is only logged. If the error is important for the caller,
// use TriggerErr instead.
func (h *ProjectionHandler) Trigger(ctx context.Context, instances ...string) context.Context {
	instances = triggerInstances(ctx, instances)
	ctx, err := h.TriggerErr(ctx, instances...)
	logging.OnError(err).WithFields(logrus.Fields{
		"projection":  h.ProjectionName,
		"instanceIDs": instances,
	}).Error("trigger failed")
	return ctx
}

// TriggerErr handles all events for the provided instances (or current instance from context if non specified)
// by calling FetchEvents and Process until the amount of events is smaller than the BulkLimit.
// If a bulk action was executed, the call timestamp in context will be reset for subsequent queries.
// The returned context is never nil. It is either the original context or an updated context.
func (h *ProjectionHandler) TriggerErr(ctx context.Context, instances ...string) (outCtx context.Context, err error) {
	instances = triggerInstances(ctx, instances)
	defer func() {
		outCtx = call.ResetTimestamp(ctx)
	}()
	for {
		events, hasLimitExceeded, err := h.FetchEvents(ctx, instances...)
		if err != nil {
			return ctx, err
		}
		if len(events) == 0 {
			return ctx, nil
		}
		_, err = h.Process(ctx, events...)
		if err != nil {
			return ctx, err
		}
		if !hasLimitExceeded {
			return ctx, nil
		}
	}
}

// Process handles multiple events by reducing them to statements and updating the projection
func (h *ProjectionHandler) Process(ctx context.Context, events ...eventstore.Event) (index int, err error) {
	if len(events) == 0 {
		return 0, nil
	}
	index = -1
	statements := make([]*Statement, len(events))
	for i, event := range events {
		statements[i], err = h.reduce(event)
		if err != nil {
			return index, err
		}
	}
	for retry := 0; retry <= h.retries; retry++ {
		index, err = h.update(ctx, statements[index+1:], h.reduce)
		if err != nil && !errors.Is(err, ErrSomeStmtsFailed) {
			return index, err
		}
		if err == nil {
			return index, nil
		}
		time.Sleep(h.retryFailedAfter)
	}
	return index, err
}

// FetchEvents checks the current sequences and filters for newer events
func (h *ProjectionHandler) FetchEvents(ctx context.Context, instances ...string) ([]eventstore.Event, bool, error) {
	if h.reduceScheduledPseudoEvent {
		return h.fetchPseudoEvents(ctx, instances...)
	}
	return h.fetchDBEvents(ctx, instances...)
}

func (h *ProjectionHandler) fetchDBEvents(ctx context.Context, instances ...string) ([]eventstore.Event, bool, error) {
	eventQuery, eventsLimit, err := h.searchQuery(ctx, instances)
	if err != nil {
		return nil, false, err
	}
	events, err := h.Eventstore.Filter(ctx, eventQuery)
	if err != nil {
		return nil, false, err
	}
	return events, int(eventsLimit) == len(events), err
}

func (h *ProjectionHandler) fetchPseudoEvents(ctx context.Context, instances ...string) ([]eventstore.Event, bool, error) {
	return []eventstore.Event{pseudo.NewScheduledEvent(ctx, time.Now(), instances...)}, false, nil
}

func (h *ProjectionHandler) subscribe(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		err := recover()
		if err != nil {
			h.Handler.Unsubscribe()
			logging.WithFields("projection", h.ProjectionName).Errorf("subscription panicked: %v", err)
		}
		cancel()
	}()
	for firstEvent := range h.EventQueue {
		events := checkAdditionalEvents(h.EventQueue, firstEvent)

		index, err := h.Process(ctx, events...)
		if err != nil || index < len(events)-1 {
			logging.WithFields("projection", h.ProjectionName).WithError(err).Warn("unable to process all events from subscription")
		}
	}
}

func (h *ProjectionHandler) schedule(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		err := recover()
		if err != nil {
			logging.WithFields("projection", h.ProjectionName, "cause", err, "stack", string(debug.Stack())).Error("schedule panicked")
		}
		cancel()
	}()
	// flag if projection has been successfully executed at least once since start
	var succeededOnce bool
	var err error
	// get every instance id except empty (system)
	query := eventstore.NewSearchQueryBuilder(eventstore.ColumnsInstanceIDs).AllowTimeTravel().AddQuery().ExcludedInstanceID("")
	for range h.triggerProjection.C {
		if !succeededOnce {
			// (re)check if it has succeeded in the meantime
			succeededOnce, err = h.hasSucceededOnce(ctx)
			if err != nil {
				logging.WithFields("projection", h.ProjectionName, "err", err).
					Error("schedule could not check if projection has already succeeded once")
				h.triggerProjection.Reset(h.requeueAfter)
				continue
			}
		}
		lockCtx := ctx
		var cancelLock context.CancelFunc
		// if it still has not succeeded, lock the projection for the system
		// so that only a single scheduler does a first schedule (of every instance)
		if !succeededOnce {
			lockCtx, cancelLock = context.WithCancel(ctx)
			errs := h.lock(lockCtx, h.requeueAfter, "system")
			if err, ok := <-errs; err != nil || !ok {
				cancelLock()
				logging.WithFields("projection", h.ProjectionName).OnError(err).Debug("initial lock failed for first schedule")
				h.triggerProjection.Reset(h.requeueAfter)
				continue
			}
			go h.cancelOnErr(lockCtx, errs, cancelLock)
		}
		if succeededOnce {
			// since we have at least one successful run, we can restrict it to events not older than
			// h.handleActiveInstances (just to be sure not to miss an event)
			// This ensures that only instances with recent events on the handler are projected
			query = query.CreationDateAfter(h.nowFunc().Add(-1 * h.handleActiveInstances))
		}
		ids, err := h.Eventstore.InstanceIDs(ctx, query.Builder())
		if err != nil {
			logging.WithFields("projection", h.ProjectionName).WithError(err).Error("instance ids")
			h.triggerProjection.Reset(h.requeueAfter)
			continue
		}
		var failed bool
		for i := 0; i < len(ids); i = i + h.concurrentInstances {
			max := i + h.concurrentInstances
			if max > len(ids) {
				max = len(ids)
			}
			instances := ids[i:max]
			lockInstanceCtx, cancelInstanceLock := context.WithCancel(lockCtx)
			errs := h.lock(lockInstanceCtx, h.requeueAfter, instances...)
			//wait until projection is locked
			if err, ok := <-errs; err != nil || !ok {
				cancelInstanceLock()
				logging.WithFields("projection", h.ProjectionName).OnError(err).Debug("initial lock failed")
				failed = true
				continue
			}
			go h.cancelOnErr(lockInstanceCtx, errs, cancelInstanceLock)
			_, err = h.TriggerErr(lockInstanceCtx, instances...)
			if err != nil {
				logging.WithFields("projection", h.ProjectionName, "instanceIDs", instances).WithError(err).Error("trigger failed")
				failed = true
			}

			cancelInstanceLock()
			unlockErr := h.unlock(instances...)
			logging.WithFields("projection", h.ProjectionName).OnError(unlockErr).Warn("unable to unlock")
		}
		// if the first schedule did not fail, store that in the eventstore, so we can check on later starts
		if !succeededOnce {
			if !failed {
				err = h.setSucceededOnce(ctx)
				logging.WithFields("projection", h.ProjectionName).OnError(err).Warn("unable to push first schedule succeeded")
			}
			cancelLock()
			unlockErr := h.unlock("system")
			logging.WithFields("projection", h.ProjectionName).OnError(unlockErr).Warn("unable to unlock first schedule")
		}
		// it succeeded at least once if it has succeeded before or if it has succeeded now - not failed ;-)
		succeededOnce = succeededOnce || !failed
		h.triggerProjection.Reset(h.requeueAfter)
	}
}

func (h *ProjectionHandler) hasSucceededOnce(ctx context.Context) (bool, error) {
	events, err := h.Eventstore.Filter(ctx, eventstore.NewSearchQueryBuilder(eventstore.ColumnsEvent).
		AddQuery().
		AggregateTypes(aggregateType).
		AggregateIDs(aggregateID).
		EventTypes(schedulerSucceeded).
		EventData(map[string]interface{}{
			"name": h.ProjectionName,
		}).
		Builder(),
	)
	return len(events) > 0 && err == nil, err
}

func (h *ProjectionHandler) setSucceededOnce(ctx context.Context) error {
	_, err := h.Eventstore.Push(ctx, &ProjectionSucceededEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(ctx,
			eventstore.NewAggregate(ctx, aggregateID, aggregateType, "v1"),
			schedulerSucceeded,
		),
		Name: h.ProjectionName,
	})
	return err
}

type ProjectionSucceededEvent struct {
	eventstore.BaseEvent `json:"-"`
	Name                 string `json:"name"`
}

func (p *ProjectionSucceededEvent) Data() interface{} {
	return p
}

func (p *ProjectionSucceededEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return nil
}

func (h *ProjectionHandler) cancelOnErr(ctx context.Context, errs <-chan error, cancel func()) {
	for {
		select {
		case err := <-errs:
			if err != nil {
				logging.WithFields("projection", h.ProjectionName).WithError(err).Warn("bulk canceled")
				cancel()
				return
			}
		case <-ctx.Done():
			cancel()
			return
		}
	}
}

func checkAdditionalEvents(eventQueue chan eventstore.Event, event eventstore.Event) []eventstore.Event {
	events := make([]eventstore.Event, 1)
	events[0] = event
	for {
		select {
		case event := <-eventQueue:
			events = append(events, event)
		default:
			return events
		}
	}
}
