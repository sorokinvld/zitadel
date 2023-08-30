package crdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/repository"
	es_repo_mock "github.com/zitadel/zitadel/internal/eventstore/repository/mock"
	"github.com/zitadel/zitadel/internal/id"
	"github.com/zitadel/zitadel/internal/repository/pseudo"
)

var (
	errFilter = errors.New("filter err")
	errReduce = errors.New("reduce err")
)

var _ eventstore.Event = &testEvent{}

type testEvent struct {
	eventstore.BaseEvent
	sequence         uint64
	previousSequence uint64
	aggregateType    eventstore.AggregateType
	instanceID       string
}

func (e *testEvent) Sequence() uint64 {
	return e.sequence
}

func (e *testEvent) Aggregate() eventstore.Aggregate {
	return eventstore.Aggregate{
		Type:       e.aggregateType,
		InstanceID: e.instanceID,
	}
}

func (e *testEvent) PreviousAggregateTypeSequence() uint64 {
	return e.previousSequence
}

func TestProjectionHandler_SearchQuery(t *testing.T) {
	type want struct {
		SearchQueryBuilder *eventstore.SearchQueryBuilder
		limit              uint64
		isErr              func(error) bool
		expectations       []mockExpectation
	}
	type fields struct {
		sequenceTable  string
		projectionName string
		reducers       []handler.AggregateReducer
		bulkLimit      uint64
	}
	type args struct {
		instanceIDs []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "error in current sequence",
			fields: fields{
				sequenceTable:  "my_sequences",
				projectionName: "my_projection",
				reducers:       failingAggregateReducers("testAgg"),
				bulkLimit:      5,
			},
			args: args{
				instanceIDs: []string{"instanceID1"},
			},
			want: want{
				limit: 0,
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrTxDone)
				},
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequenceErr(false, "my_sequences", "my_projection", []string{"instanceID1"}, sql.ErrTxDone),
					expectRollback(),
				},
				SearchQueryBuilder: nil,
			},
		},
		{
			name: "only aggregates",
			fields: fields{
				sequenceTable:  "my_sequences",
				projectionName: "my_projection",
				reducers:       failingAggregateReducers("testAgg"),
				bulkLimit:      5,
			},
			args: args{
				instanceIDs: []string{"instanceID1"},
			},
			want: want{
				limit: 5,
				isErr: func(err error) bool {
					return err == nil
				},
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID1"}),
					expectCommit(),
				},
				SearchQueryBuilder: eventstore.
					NewSearchQueryBuilder(eventstore.ColumnsEvent).
					AllowTimeTravel().
					AddQuery().
					AggregateTypes("testAgg").
					SequenceGreater(5).
					InstanceID("instanceID1").
					Builder().
					Limit(5),
			},
		},
		{
			name: "multiple instances",
			fields: fields{
				sequenceTable:  "my_sequences",
				projectionName: "my_projection",
				reducers:       failingAggregateReducers("testAgg"),
				bulkLimit:      5,
			},
			args: args{
				instanceIDs: []string{"instanceID1", "instanceID2"},
			},
			want: want{
				limit: 5,
				isErr: func(err error) bool {
					return err == nil
				},
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID1", "instanceID2"}),
					expectCommit(),
				},
				SearchQueryBuilder: eventstore.
					NewSearchQueryBuilder(eventstore.ColumnsEvent).
					AllowTimeTravel().
					AddQuery().
					AggregateTypes("testAgg").
					SequenceGreater(5).
					InstanceID("instanceID1").
					Or().
					AggregateTypes("testAgg").
					SequenceGreater(5).
					InstanceID("instanceID2").
					Builder().
					Limit(5),
			},
		},
		{
			name: "scheduled pseudo event",
			fields: fields{
				sequenceTable:  "my_sequences",
				projectionName: "my_projection",
				reducers: []handler.AggregateReducer{{
					Aggregate: pseudo.AggregateType,
					EventRedusers: []handler.EventReducer{
						{
							Event:  pseudo.ScheduledEventType,
							Reduce: testReduceErr(errors.New("should not be called")),
						},
					},
				}},
				bulkLimit: 5,
			},
			args: args{
				instanceIDs: []string{"instanceID1", "instanceID2"},
			},
			want: want{
				limit: 1,
				isErr: func(err error) bool {
					return err == nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			id.Configure(&id.Config{Identification: id.Identification{PrivateIp: id.PrivateIp{Enabled: true}}})
			h := NewStatementHandler(context.Background(), StatementHandlerConfig{
				ProjectionHandlerConfig: handler.ProjectionHandlerConfig{
					ProjectionName: tt.fields.projectionName,
				},
				SequenceTable: tt.fields.sequenceTable,
				BulkLimit:     tt.fields.bulkLimit,
				Client: &database.DB{
					DB: client,
				},
				Reducers: tt.fields.reducers,
			})

			for _, expectation := range tt.want.expectations {
				expectation(mock)
			}

			query, limit, err := h.searchQuery(context.Background(), tt.args.instanceIDs)
			if !tt.want.isErr(err) {
				t.Errorf("ProjectionHandler.prepareBulkStmts() error = %v", err)
				return
			}

			if !reflect.DeepEqual(query, tt.want.SearchQueryBuilder) {
				t.Errorf("unexpected query: expected %v, got %v", tt.want.SearchQueryBuilder, query)
			}
			if tt.want.limit != limit {
				t.Errorf("unexpected limit: got: %d want %d", limit, tt.want.limit)
			}

			mock.MatchExpectationsInOrder(true)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations not met: %v", err)
			}
		})
	}
}

func TestStatementHandler_Update(t *testing.T) {
	type fields struct {
		eventstore *eventstore.Eventstore
		aggregates []eventstore.AggregateType
	}
	type want struct {
		expectations []mockExpectation
		isErr        func(error) bool
		stmtsLen     int
	}
	type args struct {
		ctx    context.Context
		stmts  []*handler.Statement
		reduce handler.Reduce
	}
	tests := []struct {
		name   string
		fields fields
		want   want
		args   args
	}{
		{
			name: "begin fails",
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewNoOpStatement(&testEvent{
						aggregateType:    "agg",
						sequence:         6,
						previousSequence: 0,
					}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBeginErr(sql.ErrConnDone),
				},
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrConnDone)
				},
			},
		},
		{
			name: "current sequence fails",
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewNoOpStatement(&testEvent{
						aggregateType:    "agg",
						sequence:         6,
						previousSequence: 0,
						instanceID:       "instanceID",
					}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequenceErr(false, "my_sequences", "my_projection", []string{"instanceID"}, sql.ErrTxDone),
					expectRollback(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrTxDone)
				},
			},
		},
		{
			name: "fetch previous fails",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).
						ExpectFilterEventsError(errFilter),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewNoOpStatement(&testEvent{
						aggregateType:    "agg",
						sequence:         6,
						previousSequence: 0,
						instanceID:       "instanceID",
					}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID"}),
					expectRollback(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, errFilter)
				},
				stmtsLen: 1,
			},
		},
		{
			name: "no successful stmts",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "testAgg",
							sequence:         7,
							previousSequence: 6,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID"}),
					expectCommit(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, handler.ErrSomeStmtsFailed)
				},
				stmtsLen: 1,
			},
		},
		{
			name: "update current sequence fails",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t),
				),
				),
				aggregates: []eventstore.AggregateType{"agg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         7,
							previousSequence: 5,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "agg", []string{"instanceID"}),
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
					expectUpdateCurrentSequenceNoRows("my_sequences", "my_projection", 7, "agg", "instanceID"),
					expectRollback(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, errSeqNotUpdated)
				},
				stmtsLen: 1,
			},
		},
		{
			name: "commit fails",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t),
				),
				),
				aggregates: []eventstore.AggregateType{"agg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         7,
							previousSequence: 5,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "agg", []string{"instanceID"}),
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
					expectUpdateCurrentSequence("my_sequences", "my_projection", 7, "agg", "instanceID"),
					expectCommitErr(sql.ErrConnDone),
				},
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrConnDone)
				},
				stmtsLen: 1,
			},
		},
		{
			name: "correct",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewNoOpStatement(&testEvent{
						aggregateType:    "testAgg",
						sequence:         7,
						previousSequence: 5,
						instanceID:       "instanceID",
					}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID"}),
					expectUpdateCurrentSequence("my_sequences", "my_projection", 7, "testAgg", "instanceID"),
					expectCommit(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
				stmtsLen: 0,
			},
		},
		{
			name: "fetch previous stmts no additional stmts",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).ExpectFilterEvents(),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewNoOpStatement(&testEvent{
						aggregateType:    "testAgg",
						sequence:         7,
						previousSequence: 0,
						instanceID:       "instanceID",
					}),
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID"}),
					expectUpdateCurrentSequence("my_sequences", "my_projection", 7, "testAgg", "instanceID"),
					expectCommit(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
			},
		},
		{
			name: "fetch previous stmts additional events",
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).ExpectFilterEvents(
						&repository.Event{
							AggregateType:             "testAgg",
							Sequence:                  6,
							PreviousAggregateSequence: 5,
							InstanceID:                "instanceID",
						},
					),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			args: args{
				ctx: context.Background(),
				stmts: []*handler.Statement{
					NewNoOpStatement(&testEvent{
						aggregateType:    "testAgg",
						sequence:         7,
						previousSequence: 0,
						instanceID:       "instanceID",
					}),
				},
				reduce: testReduce(),
			},
			want: want{
				expectations: []mockExpectation{
					expectBegin(),
					expectCurrentSequence(false, "my_sequences", "my_projection", 5, "testAgg", []string{"instanceID"}),
					expectUpdateCurrentSequence("my_sequences", "my_projection", 7, "testAgg", "instanceID"),
					expectCommit(),
				},
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
				stmtsLen: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			h := &StatementHandler{
				ProjectionHandler: &handler.ProjectionHandler{
					Handler: handler.Handler{
						Eventstore: tt.fields.eventstore,
					},
					ProjectionName: "my_projection",
				},
				sequenceTable:           "my_sequences",
				currentSequenceStmt:     fmt.Sprintf(currentSequenceStmtFormat, "my_sequences"),
				updateSequencesBaseStmt: fmt.Sprintf(updateCurrentSequencesStmtFormat, "my_sequences"),
				client: &database.DB{
					DB: client,
				},
			}

			h.aggregates = tt.fields.aggregates

			for _, expectation := range tt.want.expectations {
				expectation(mock)
			}

			index, err := h.Update(tt.args.ctx, tt.args.stmts, tt.args.reduce)
			if !tt.want.isErr(err) {
				t.Errorf("StatementHandler.Update() error = %v", err)
			}
			if err == nil && tt.want.stmtsLen != index {
				t.Errorf("wrong stmts length: want: %d got %d", tt.want.stmtsLen, index)
			}

			mock.MatchExpectationsInOrder(true)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations not met: %v", err)
			}
		})
	}
}

func TestProjectionHandler_fetchPreviousStmts(t *testing.T) {
	type args struct {
		ctx       context.Context
		stmtSeq   uint64
		sequences currentSequences
		reduce    handler.Reduce
	}
	type want struct {
		stmtCount int
		isErr     func(error) bool
	}
	type fields struct {
		eventstore *eventstore.Eventstore
		aggregates []eventstore.AggregateType
	}
	tests := []struct {
		name   string
		args   args
		fields fields
		want   want
	}{
		{
			name: "no queries",
			args: args{
				ctx:    context.Background(),
				reduce: testReduce(),
			},
			fields: fields{
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
				stmtCount: 0,
			},
		},
		{
			name: "eventstore returns err",
			args: args{
				ctx:    context.Background(),
				reduce: testReduce(),
				sequences: currentSequences{
					"testAgg": []*instanceSequence{
						{sequence: 5},
					},
				},
				stmtSeq: 6,
			},
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).ExpectFilterEventsError(errFilter),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, errFilter)
				},
				stmtCount: 0,
			},
		},
		{
			name: "no events found",
			args: args{
				ctx:    context.Background(),
				reduce: testReduce(),
				sequences: currentSequences{
					"testAgg": []*instanceSequence{
						{sequence: 5},
					},
				},
				stmtSeq: 6,
			},
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).ExpectFilterEvents(),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			want: want{
				isErr: func(err error) bool {
					return err == nil
				},
			},
		},
		{
			name: "found events",
			args: args{
				ctx:    context.Background(),
				reduce: testReduce(),
				sequences: currentSequences{
					"testAgg": []*instanceSequence{
						{sequence: 5},
					},
				},
				stmtSeq: 10,
			},
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).ExpectFilterEvents(
						&repository.Event{
							ID:                        "id",
							Sequence:                  7,
							PreviousAggregateSequence: 0,
							CreationDate:              time.Now(),
							Type:                      "test.added",
							Version:                   "v1",
							AggregateID:               "testid",
							AggregateType:             "testAgg",
						},
						&repository.Event{
							ID:                        "id",
							Sequence:                  9,
							PreviousAggregateSequence: 7,
							CreationDate:              time.Now(),
							Type:                      "test.changed",
							Version:                   "v1",
							AggregateID:               "testid",
							AggregateType:             "testAgg",
						},
					),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			want: want{
				stmtCount: 2,
				isErr: func(err error) bool {
					return err == nil
				},
			},
		},
		{
			name: "reduce fails",
			args: args{
				ctx:    context.Background(),
				reduce: testReduceErr(errReduce),
				sequences: currentSequences{
					"testAgg": []*instanceSequence{
						{sequence: 5},
					},
				},
				stmtSeq: 10,
			},
			fields: fields{
				eventstore: eventstore.NewEventstore(eventstore.TestConfig(
					es_repo_mock.NewRepo(t).ExpectFilterEvents(
						&repository.Event{
							ID:                        "id",
							Sequence:                  7,
							PreviousAggregateSequence: 0,
							CreationDate:              time.Now(),
							Type:                      "test.added",
							Version:                   "v1",
							AggregateID:               "testid",
							AggregateType:             "testAgg",
						},
					),
				),
				),
				aggregates: []eventstore.AggregateType{"testAgg"},
			},
			want: want{
				stmtCount: 0,
				isErr: func(err error) bool {
					return errors.Is(err, errReduce)
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatementHandler{
				aggregates: tt.fields.aggregates,
			}
			h.ProjectionHandler = &handler.ProjectionHandler{
				Handler: handler.Handler{
					Eventstore: tt.fields.eventstore,
				},
				ProjectionName: "my_projection",
			}
			stmts, err := h.fetchPreviousStmts(tt.args.ctx, nil, tt.args.stmtSeq, "", tt.args.sequences, tt.args.reduce)
			if !tt.want.isErr(err) {
				t.Errorf("ProjectionHandler.prepareBulkStmts() error = %v", err)
				return
			}
			if tt.want.stmtCount != len(stmts) {
				t.Errorf("unexpected length of stmts: got: %d want %d", len(stmts), tt.want.stmtCount)
			}
		})
	}
}

func TestStatementHandler_executeStmts(t *testing.T) {
	type fields struct {
		projectionName    string
		maxFailureCount   uint
		failedEventsTable string
	}
	type args struct {
		stmts     []*handler.Statement
		sequences currentSequences
	}
	type want struct {
		expectations []mockExpectation
		idx          int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "already inserted",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         5,
							previousSequence: 2,
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val1",
							},
						}),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{sequence: 5},
					},
				},
			},
			want: want{
				expectations: []mockExpectation{},
				idx:          -1,
			},
		},
		{
			name: "previous sequence higher than sequence",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         5,
							previousSequence: 0,
						},
						[]handler.Column{
							{
								Name:  "col1",
								Value: "val1",
							},
						}),
					NewCreateStatement(

						&testEvent{
							aggregateType:    "agg",
							sequence:         8,
							previousSequence: 7,
						},
						[]handler.Column{
							{
								Name:  "col2",
								Value: "val2",
							},
						}),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{sequence: 2},
					},
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreate("my_projection", []string{"col1"}, []string{"$1"}),
					expectSavePointRelease(),
				},
				idx: 0,
			},
		},
		{
			name: "execute fails not continue",
			fields: fields{
				projectionName:    "my_projection",
				maxFailureCount:   5,
				failedEventsTable: "failed_events",
			},
			args: args{
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         5,
							previousSequence: 0,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         6,
							previousSequence: 5,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         7,
							previousSequence: 6,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{sequence: 2},
					},
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
					expectSavePoint(),
					expectCreateErr("my_projection", []string{"col"}, []string{"$1"}, sql.ErrConnDone),
					expectSavePointRollback(),
					expectFailureCount("failed_events", "my_projection", "instanceID", 6, 3),
					expectUpdateFailureCount("failed_events", "my_projection", "instanceID", 6, 4),
				},
				idx: 0,
			},
		},
		{
			name: "execute fails continue",
			fields: fields{
				projectionName:    "my_projection",
				maxFailureCount:   5,
				failedEventsTable: "failed_events",
			},
			args: args{
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         5,
							previousSequence: 0,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col1",
								Value: "val1",
							},
						}),
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         6,
							previousSequence: 5,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col2",
								Value: "val2",
							},
						}),
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         7,
							previousSequence: 6,
							instanceID:       "instanceID",
						},
						[]handler.Column{
							{
								Name:  "col3",
								Value: "val3",
							},
						}),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{sequence: 2},
					},
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreate("my_projection", []string{"col1"}, []string{"$1"}),
					expectSavePointRelease(),
					expectSavePoint(),
					expectCreateErr("my_projection", []string{"col2"}, []string{"$1"}, sql.ErrConnDone),
					expectSavePointRollback(),
					expectFailureCount("failed_events", "my_projection", "instanceID", 6, 4),
					expectUpdateFailureCount("failed_events", "my_projection", "instanceID", 6, 5),
					expectSavePoint(),
					expectCreate("my_projection", []string{"col3"}, []string{"$1"}),
					expectSavePointRelease(),
				},
				idx: 2,
			},
		},
		{
			name: "correct",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmts: []*handler.Statement{
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         5,
							previousSequence: 0,
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         6,
							previousSequence: 5,
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
					NewCreateStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         7,
							previousSequence: 6,
						},
						[]handler.Column{
							{
								Name:  "col",
								Value: "val",
							},
						}),
					NewMultiStatement(
						&testEvent{
							aggregateType:    "agg",
							sequence:         8,
							previousSequence: 7,
						},
						AddCreateStatement(
							[]handler.Column{
								{
									Name:  "col",
									Value: "val",
								},
							},
						),
						AddCreateStatement(
							[]handler.Column{
								{
									Name:  "col",
									Value: "val",
								},
							},
						),
					),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{sequence: 2},
					},
				},
			},
			want: want{
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
				},
				idx: 3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			h := NewStatementHandler(
				context.Background(),
				StatementHandlerConfig{
					ProjectionHandlerConfig: handler.ProjectionHandlerConfig{
						HandlerConfig: handler.HandlerConfig{
							Eventstore: nil,
						},
						ProjectionName: tt.fields.projectionName,
						RequeueEvery:   0,
					},
					Client: &database.DB{
						DB: client,
					},
					FailedEventsTable: tt.fields.failedEventsTable,
					MaxFailureCount:   tt.fields.maxFailureCount,
				},
			)

			mock.ExpectBegin()

			for _, expectation := range tt.want.expectations {
				expectation(mock)
			}

			mock.ExpectCommit()

			tx, err := client.Begin()
			if err != nil {
				t.Fatalf("unexpected err in begin: %v", err)
			}

			idx := h.executeStmts(tx, &tt.args.stmts, tt.args.sequences)
			if idx != tt.want.idx {
				t.Errorf("unexpected index want: %d got %d", tt.want.idx, idx)
			}

			if err := tx.Commit(); err != nil {
				t.Fatalf("unexpected err in commit: %v", err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations not met: %v", err)
			}
		})
	}
}

func TestStatementHandler_executeStmt(t *testing.T) {
	type fields struct {
		projectionName string
	}
	type args struct {
		stmt *handler.Statement
	}
	type want struct {
		expectations []mockExpectation
		isErr        func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "create savepoint fails",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmt: NewCreateStatement(
					&testEvent{
						aggregateType:    "agg",
						sequence:         1,
						previousSequence: 0,
					},
					[]handler.Column{
						{
							Name:  "col",
							Value: "val",
						},
					}),
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrConnDone)
				},
				expectations: []mockExpectation{
					expectSavePointErr(sql.ErrConnDone),
				},
			},
		},
		{
			name: "execute fails",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmt: NewCreateStatement(
					&testEvent{
						aggregateType:    "agg",
						sequence:         1,
						previousSequence: 0,
					},
					[]handler.Column{
						{
							Name:  "col",
							Value: "val",
						},
					}),
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrNoRows)
				},
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreateErr("my_projection", []string{"col"}, []string{"$1"}, sql.ErrNoRows),
					expectSavePointRollback(),
				},
			},
		},
		{
			name: "rollback savepoint fails",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmt: NewCreateStatement(
					&testEvent{
						aggregateType:    "agg",
						sequence:         1,
						previousSequence: 0,
					},
					[]handler.Column{
						{
							Name:  "col",
							Value: "val",
						},
					}),
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrConnDone)
				},
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreateErr("my_projection", []string{"col"}, []string{"$1"}, sql.ErrNoRows),
					expectSavePointRollbackErr(sql.ErrConnDone),
				},
			},
		},
		{
			name: "no op",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmt: NewNoOpStatement(&testEvent{
					aggregateType:    "agg",
					sequence:         1,
					previousSequence: 0,
				}),
			},
			want: want{
				isErr: func(err error) bool {
					return err == nil
				},
				expectations: []mockExpectation{},
			},
		},
		{
			name: "with op",
			fields: fields{
				projectionName: "my_projection",
			},
			args: args{
				stmt: NewCreateStatement(
					&testEvent{
						aggregateType:    "agg",
						sequence:         1,
						previousSequence: 0,
					},
					[]handler.Column{
						{
							Name:  "col",
							Value: "val",
						},
					}),
			},
			want: want{
				isErr: func(err error) bool {
					return err == nil
				},
				expectations: []mockExpectation{
					expectSavePoint(),
					expectCreate("my_projection", []string{"col"}, []string{"$1"}),
					expectSavePointRelease(),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatementHandler{
				ProjectionHandler: &handler.ProjectionHandler{
					ProjectionName: tt.fields.projectionName,
				},
			}

			client, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			mock.ExpectBegin()

			for _, expectation := range tt.want.expectations {
				expectation(mock)
			}

			mock.ExpectCommit()

			tx, err := client.Begin()
			if err != nil {
				t.Fatalf("unexpected err in begin: %v", err)
			}

			err = h.executeStmt(tx, tt.args.stmt)
			if !tt.want.isErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if err = tx.Commit(); err != nil {
				t.Fatalf("unexpected err in begin: %v", err)
			}

			mock.MatchExpectationsInOrder(true)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations not met: %v", err)
			}
		})
	}
}

func TestStatementHandler_currentSequence(t *testing.T) {
	type fields struct {
		sequenceTable  string
		projectionName string
		aggregates     []eventstore.AggregateType
	}
	type args struct {
		stmt        handler.Statement
		instanceIDs []string
	}
	type want struct {
		expectations []mockExpectation
		isErr        func(error) bool
		sequences    currentSequences
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "error in query",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
			},
			args: args{
				stmt: handler.Statement{},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrConnDone)
				},
				expectations: []mockExpectation{
					expectCurrentSequenceErr(true, "my_table", "my_projection", nil, sql.ErrConnDone),
				},
			},
		},
		{
			name: "no rows",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				stmt:        handler.Statement{},
				instanceIDs: []string{"instanceID"},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
				expectations: []mockExpectation{
					expectCurrentSequenceNoRows("my_table", "my_projection", []string{"instanceID"}),
				},
				sequences: currentSequences{},
			},
		},
		{
			name: "scan err",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				stmt:        handler.Statement{},
				instanceIDs: []string{"instanceID"},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrTxDone)
				},
				expectations: []mockExpectation{
					expectCurrentSequenceScanErr("my_table", "my_projection", []string{"instanceID"}),
				},
				sequences: currentSequences{},
			},
		},
		{
			name: "found",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				stmt:        handler.Statement{},
				instanceIDs: []string{"instanceID"},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
				expectations: []mockExpectation{
					expectCurrentSequence(true, "my_table", "my_projection", 5, "agg", []string{"instanceID"}),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{
							sequence:   5,
							instanceID: "instanceID",
						},
					},
				},
			},
		},
		{
			name: "multiple found",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				stmt:        handler.Statement{},
				instanceIDs: []string{"instanceID1", "instanceID2"},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, nil)
				},
				expectations: []mockExpectation{
					expectCurrentSequence(true, "my_table", "my_projection", 5, "agg", []string{"instanceID1", "instanceID2"}),
				},
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{
							sequence:   5,
							instanceID: "instanceID1",
						},
						{
							sequence:   5,
							instanceID: "instanceID2",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StatementHandler{
				ProjectionHandler: &handler.ProjectionHandler{
					ProjectionName: tt.fields.projectionName,
				},
				sequenceTable:       tt.fields.sequenceTable,
				currentSequenceStmt: fmt.Sprintf(currentSequenceStmtFormat, tt.fields.sequenceTable),
			}

			h.aggregates = tt.fields.aggregates

			client, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			mock.ExpectBegin()

			for _, expectation := range tt.want.expectations {
				expectation(mock)
			}

			mock.ExpectCommit()

			tx, err := client.Begin()
			if err != nil {
				t.Fatalf("unexpected err in begin: %v", err)
			}

			seq, err := h.currentSequences(context.Background(), true, (&transaction{Tx: tx}).QueryContext, tt.args.instanceIDs)
			if !tt.want.isErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			if err = tx.Commit(); err != nil {
				t.Fatalf("unexpected err in commit: %v", err)
			}

			mock.MatchExpectationsInOrder(true)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations not met: %v", err)
			}

			for _, aggregateType := range tt.fields.aggregates {
				assert.Equal(t, tt.want.sequences[aggregateType], seq[aggregateType])
			}
		})
	}
}

func TestStatementHandler_updateCurrentSequence(t *testing.T) {
	type fields struct {
		sequenceTable  string
		projectionName string
		aggregates     []eventstore.AggregateType
	}
	type args struct {
		sequences currentSequences
	}
	type want struct {
		expectations []mockExpectation
		isErr        func(error) bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "update sequence fails",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{
							sequence:   5,
							instanceID: "instanceID",
						},
					},
				},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, sql.ErrConnDone)
				},
				expectations: []mockExpectation{
					expectUpdateCurrentSequenceErr("my_table", "my_projection", 5, sql.ErrConnDone, "agg", "instanceID"),
				},
			},
		},
		{
			name: "update sequence returns no rows",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{
							sequence:   5,
							instanceID: "instanceID",
						},
					},
				},
			},
			want: want{
				isErr: func(err error) bool {
					return errors.Is(err, errSeqNotUpdated)
				},
				expectations: []mockExpectation{
					expectUpdateCurrentSequenceNoRows("my_table", "my_projection", 5, "agg", "instanceID"),
				},
			},
		},
		{
			name: "correct",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{
							sequence:   5,
							instanceID: "instanceID",
						},
					},
				},
			},
			want: want{
				isErr: func(err error) bool {
					return err == nil
				},
				expectations: []mockExpectation{
					expectUpdateCurrentSequence("my_table", "my_projection", 5, "agg", "instanceID"),
				},
			},
		},
		{
			name: "multiple sequences",
			fields: fields{
				sequenceTable:  "my_table",
				projectionName: "my_projection",
				aggregates:     []eventstore.AggregateType{"agg"},
			},
			args: args{
				sequences: currentSequences{
					"agg": []*instanceSequence{
						{
							sequence:   5,
							instanceID: "instanceID",
						},
					},
					"agg2": []*instanceSequence{
						{
							sequence:   6,
							instanceID: "instanceID",
						},
						{
							sequence:   10,
							instanceID: "instanceID2",
						},
					},
				},
			},
			want: want{
				isErr: func(err error) bool {
					return err == nil
				},
				expectations: []mockExpectation{
					expectUpdateThreeCurrentSequence(t, "my_table", "my_projection", currentSequences{
						"agg": []*instanceSequence{
							{
								sequence:   5,
								instanceID: "instanceID",
							},
						},
						"agg2": []*instanceSequence{
							{
								sequence:   6,
								instanceID: "instanceID",
							},
							{
								sequence:   10,
								instanceID: "instanceID2",
							},
						},
					}),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			h := &StatementHandler{
				ProjectionHandler: &handler.ProjectionHandler{
					ProjectionName: tt.fields.projectionName,
				},
				sequenceTable:           tt.fields.sequenceTable,
				updateSequencesBaseStmt: fmt.Sprintf(updateCurrentSequencesStmtFormat, tt.fields.sequenceTable),
			}

			h.aggregates = tt.fields.aggregates

			client, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer client.Close()

			mock.ExpectBegin()
			for _, expectation := range tt.want.expectations {
				expectation(mock)
			}
			mock.ExpectCommit()

			tx, err := client.Begin()
			if err != nil {
				t.Fatalf("unexpected error in begin: %v", err)
			}

			err = h.updateCurrentSequences(tx, tt.args.sequences)
			if !tt.want.isErr(err) {
				t.Errorf("unexpected error: %v", err)
			}

			err = tx.Commit()
			if err != nil {
				t.Fatalf("unexpected error in commit: %v", err)
			}

			mock.MatchExpectationsInOrder(true)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("expectations not met: %v", err)
			}
		})
	}
}

func testReduce() handler.Reduce {
	return func(event eventstore.Event) (*handler.Statement, error) {
		return NewNoOpStatement(event), nil
	}
}

func testReduceErr(err error) handler.Reduce {
	return func(event eventstore.Event) (*handler.Statement, error) {
		return nil, err
	}
}

func failingAggregateReducers(aggregates ...eventstore.AggregateType) []handler.AggregateReducer {
	reducers := make([]handler.AggregateReducer, len(aggregates))
	for idx := range aggregates {
		reducers[idx] = handler.AggregateReducer{
			Aggregate: aggregates[idx],
			EventRedusers: []handler.EventReducer{{
				Event:  "any.event",
				Reduce: testReduceErr(errors.New("should not be called")),
			}},
		}
	}
	return reducers
}
