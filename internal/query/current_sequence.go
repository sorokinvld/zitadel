package query

import (
	"context"
	"database/sql"
	errs "errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/api/call"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/query/projection"
	"github.com/zitadel/zitadel/internal/telemetry/tracing"
)

const (
	lockStmtFormat = "UPDATE %[1]s" +
		" set locker_id = $1, locked_until = now()+$2::INTERVAL" +
		" WHERE projection_name = $3"
	lockerIDReset = "reset"
)

type LatestSequence struct {
	Sequence  uint64
	Timestamp time.Time
}

type CurrentSequences struct {
	SearchResponse
	CurrentSequences []*CurrentSequence
}

type CurrentSequence struct {
	ProjectionName  string
	CurrentSequence uint64
	Timestamp       time.Time
}

type CurrentSequencesSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

func NewCurrentSequencesInstanceIDSearchQuery(instanceID string) (SearchQuery, error) {
	return NewTextQuery(CurrentSequenceColInstanceID, instanceID, TextEquals)
}

func (q *CurrentSequencesSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

func (q *Queries) SearchCurrentSequences(ctx context.Context, queries *CurrentSequencesSearchQueries) (failedEvents *CurrentSequences, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareCurrentSequencesQuery(ctx, q.client)
	stmt, args, err := queries.toQuery(query).ToSql()
	if err != nil {
		return nil, errors.ThrowInvalidArgument(err, "QUERY-MmFef", "Errors.Query.InvalidRequest")
	}

	err = q.client.QueryContext(ctx, func(rows *sql.Rows) error {
		failedEvents, err = scan(rows)
		return err
	}, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-22H8f", "Errors.Internal")
	}
	return failedEvents, nil
}

func (q *Queries) latestSequence(ctx context.Context, projections ...table) (seq *LatestSequence, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareLatestSequence(ctx, q.client)
	or := make(sq.Or, len(projections))
	for i, projection := range projections {
		or[i] = sq.Eq{CurrentSequenceColProjectionName.identifier(): projection.name}
	}
	stmt, args, err := query.
		Where(or).
		Where(sq.Eq{CurrentSequenceColInstanceID.identifier(): authz.GetInstance(ctx).InstanceID()}).
		OrderBy(CurrentSequenceColCurrentSequence.identifier() + " DESC").
		ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-5CfX9", "Errors.Query.SQLStatement")
	}

	err = q.client.QueryRowContext(ctx, func(row *sql.Row) error {
		seq, err = scan(row)
		return err
	}, stmt, args...)

	return seq, err
}

func (q *Queries) ClearCurrentSequence(ctx context.Context, projectionName string) (err error) {
	err = q.checkAndLock(ctx, projectionName)
	if err != nil {
		return err
	}

	tx, err := q.client.Begin()
	if err != nil {
		return errors.ThrowInternal(err, "QUERY-9iOpr", "Errors.RemoveFailed")
	}
	tables, err := tablesForReset(ctx, tx, projectionName)
	if err != nil {
		return err
	}
	err = reset(tx, tables, projectionName)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (q *Queries) checkAndLock(ctx context.Context, projectionName string) error {
	projectionQuery, args, err := sq.Select("count(*)").
		From("[show tables from projections]").
		Where(
			sq.And{
				sq.NotEq{"table_name": []string{"locks", "current_sequences", "failed_events"}},
				sq.Eq{"concat('projections.', table_name)": projectionName},
			}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return errors.ThrowInternal(err, "QUERY-Dfwf2", "Errors.ProjectionName.Invalid")
	}
	var count int
	err = q.client.QueryRowContext(ctx, func(row *sql.Row) error {
		if err := row.Scan(&count); err != nil || count == 0 {
			return errors.ThrowInternal(err, "QUERY-ej8fn", "Errors.ProjectionName.Invalid")
		}
		return err
	}, projectionQuery, args...)
	if err != nil {
		return err
	}

	lock := fmt.Sprintf(lockStmtFormat, locksTable.identifier())
	if err != nil {
		return errors.ThrowInternal(err, "QUERY-DVfg3", "Errors.RemoveFailed")
	}
	//lock for twice the default duration (10s)
	res, err := q.client.ExecContext(ctx, lock, lockerIDReset, 20*time.Second, projectionName)
	if err != nil {
		return errors.ThrowInternal(err, "QUERY-WEfr2", "Errors.RemoveFailed")
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return errors.ThrowInternal(err, "QUERY-Bh3ws", "Errors.RemoveFailed")
	}
	time.Sleep(7 * time.Second) //more than half the default lock duration (10s)
	return nil
}

func tablesForReset(ctx context.Context, tx *sql.Tx, projectionName string) ([]string, error) {
	tablesQuery, args, err := sq.Select("concat('projections.', table_name)").
		From("[show tables from projections]").
		Where(
			sq.And{
				sq.Eq{"type": "table"},
				sq.NotEq{"table_name": []string{"locks", "current_sequences", "failed_events"}},
				sq.Like{"concat('projections.', table_name)": projectionName + "%"},
			}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-ASff2", "Errors.ProjectionName.Invalid")
	}
	var tables []string
	rows, err := tx.QueryContext(ctx, tablesQuery, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Dgfw", "Errors.ProjectionName.Invalid")
	}
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, errors.ThrowInternal(err, "QUERY-ej8fn", "Errors.ProjectionName.Invalid")
		}
		tables = append(tables, tableName)
	}
	return tables, nil
}

func reset(tx *sql.Tx, tables []string, projectionName string) error {
	for _, tableName := range tables {
		_, err := tx.Exec(fmt.Sprintf("TRUNCATE %s cascade", tableName))
		if err != nil {
			return errors.ThrowInternal(err, "QUERY-3n92f", "Errors.RemoveFailed")
		}
	}
	update, args, err := sq.Update(currentSequencesTable.identifier()).
		Set(CurrentSequenceColCurrentSequence.name, 0).
		Where(sq.Eq{CurrentSequenceColProjectionName.name: projectionName}).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return errors.ThrowInternal(err, "QUERY-Ff3tw", "Errors.RemoveFailed")
	}
	_, err = tx.Exec(update, args...)
	if err != nil {
		return errors.ThrowInternal(err, "QUERY-NFiws", "Errors.RemoveFailed")
	}
	return nil
}

func prepareLatestSequence(ctx context.Context, db prepareDatabase) (sq.SelectBuilder, func(*sql.Row) (*LatestSequence, error)) {
	return sq.Select(
			CurrentSequenceColCurrentSequence.identifier(),
			CurrentSequenceColTimestamp.identifier()).
			From(currentSequencesTable.identifier() + db.Timetravel(call.Took(ctx))).
			PlaceholderFormat(sq.Dollar),
		func(row *sql.Row) (*LatestSequence, error) {
			seq := new(LatestSequence)
			err := row.Scan(
				&seq.Sequence,
				&seq.Timestamp,
			)
			if err != nil && !errs.Is(err, sql.ErrNoRows) {
				return nil, errors.ThrowInternal(err, "QUERY-aAZ1D", "Errors.Internal")
			}
			return seq, nil
		}
}

func prepareCurrentSequencesQuery(ctx context.Context, db prepareDatabase) (sq.SelectBuilder, func(*sql.Rows) (*CurrentSequences, error)) {
	return sq.Select(
			"max("+CurrentSequenceColCurrentSequence.identifier()+") as "+CurrentSequenceColCurrentSequence.name,
			"max("+CurrentSequenceColTimestamp.identifier()+") as "+CurrentSequenceColTimestamp.name,
			CurrentSequenceColProjectionName.identifier(),
			countColumn.identifier()).
			From(currentSequencesTable.identifier() + db.Timetravel(call.Took(ctx))).
			GroupBy(CurrentSequenceColProjectionName.identifier()).
			PlaceholderFormat(sq.Dollar),
		func(rows *sql.Rows) (*CurrentSequences, error) {
			currentSequences := make([]*CurrentSequence, 0)
			var count uint64
			for rows.Next() {
				currentSequence := new(CurrentSequence)
				err := rows.Scan(
					&currentSequence.CurrentSequence,
					&currentSequence.Timestamp,
					&currentSequence.ProjectionName,
					&count,
				)
				if err != nil {
					return nil, err
				}
				currentSequences = append(currentSequences, currentSequence)
			}

			if err := rows.Close(); err != nil {
				return nil, errors.ThrowInternal(err, "QUERY-jbJ77", "Errors.Query.CloseRows")
			}

			return &CurrentSequences{
				CurrentSequences: currentSequences,
				SearchResponse: SearchResponse{
					Count: count,
				},
			}, nil
		}
}

var (
	currentSequencesTable = table{
		name:          projection.CurrentSeqTable,
		instanceIDCol: "instance_id",
	}
	CurrentSequenceColAggregateType = Column{
		name:  "aggregate_type",
		table: currentSequencesTable,
	}
	CurrentSequenceColCurrentSequence = Column{
		name:  "current_sequence",
		table: currentSequencesTable,
	}
	CurrentSequenceColTimestamp = Column{
		name:  "timestamp",
		table: currentSequencesTable,
	}
	CurrentSequenceColProjectionName = Column{
		name:  "projection_name",
		table: currentSequencesTable,
	}
	CurrentSequenceColInstanceID = Column{
		name:  "instance_id",
		table: currentSequencesTable,
	}
)

var (
	locksTable = table{
		name:          projection.LocksTable,
		instanceIDCol: "instance_id",
	}
	LocksColLockerID = Column{
		name:  "locker_id",
		table: locksTable,
	}
	LocksColUntil = Column{
		name:  "locked_until",
		table: locksTable,
	}
	LocksColProjectionName = Column{
		name:  "projection_name",
		table: locksTable,
	}
)
