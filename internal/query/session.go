package query

import (
	"context"
	"database/sql"
	errs "errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/api/call"
	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/query/projection"
	"github.com/zitadel/zitadel/internal/telemetry/tracing"
)

type Sessions struct {
	SearchResponse
	Sessions []*Session
}

type Session struct {
	ID             string
	CreationDate   time.Time
	ChangeDate     time.Time
	Sequence       uint64
	State          domain.SessionState
	ResourceOwner  string
	Creator        string
	UserFactor     SessionUserFactor
	PasswordFactor SessionPasswordFactor
	IntentFactor   SessionIntentFactor
	WebAuthNFactor SessionWebAuthNFactor
	TOTPFactor     SessionTOTPFactor
	OTPSMSFactor   SessionOTPFactor
	OTPEmailFactor SessionOTPFactor
	Metadata       map[string][]byte
}

type SessionUserFactor struct {
	UserID        string
	ResourceOwner string
	UserCheckedAt time.Time
	LoginName     string
	DisplayName   string
}

type SessionPasswordFactor struct {
	PasswordCheckedAt time.Time
}

type SessionIntentFactor struct {
	IntentCheckedAt time.Time
}

type SessionWebAuthNFactor struct {
	WebAuthNCheckedAt time.Time
	UserVerified      bool
}

type SessionTOTPFactor struct {
	TOTPCheckedAt time.Time
}

type SessionOTPFactor struct {
	OTPCheckedAt time.Time
}

type SessionsSearchQueries struct {
	SearchRequest
	Queries []SearchQuery
}

func (q *SessionsSearchQueries) toQuery(query sq.SelectBuilder) sq.SelectBuilder {
	query = q.SearchRequest.toQuery(query)
	for _, q := range q.Queries {
		query = q.toQuery(query)
	}
	return query
}

var (
	sessionsTable = table{
		name:          projection.SessionsProjectionTable,
		instanceIDCol: projection.SessionColumnInstanceID,
	}
	SessionColumnID = Column{
		name:  projection.SessionColumnID,
		table: sessionsTable,
	}
	SessionColumnCreationDate = Column{
		name:  projection.SessionColumnCreationDate,
		table: sessionsTable,
	}
	SessionColumnChangeDate = Column{
		name:  projection.SessionColumnChangeDate,
		table: sessionsTable,
	}
	SessionColumnSequence = Column{
		name:  projection.SessionColumnSequence,
		table: sessionsTable,
	}
	SessionColumnState = Column{
		name:  projection.SessionColumnState,
		table: sessionsTable,
	}
	SessionColumnResourceOwner = Column{
		name:  projection.SessionColumnResourceOwner,
		table: sessionsTable,
	}
	SessionColumnInstanceID = Column{
		name:  projection.SessionColumnInstanceID,
		table: sessionsTable,
	}
	SessionColumnCreator = Column{
		name:  projection.SessionColumnCreator,
		table: sessionsTable,
	}
	SessionColumnUserID = Column{
		name:  projection.SessionColumnUserID,
		table: sessionsTable,
	}
	SessionColumnUserCheckedAt = Column{
		name:  projection.SessionColumnUserCheckedAt,
		table: sessionsTable,
	}
	SessionColumnPasswordCheckedAt = Column{
		name:  projection.SessionColumnPasswordCheckedAt,
		table: sessionsTable,
	}
	SessionColumnIntentCheckedAt = Column{
		name:  projection.SessionColumnIntentCheckedAt,
		table: sessionsTable,
	}
	SessionColumnWebAuthNCheckedAt = Column{
		name:  projection.SessionColumnWebAuthNCheckedAt,
		table: sessionsTable,
	}
	SessionColumnWebAuthNUserVerified = Column{
		name:  projection.SessionColumnWebAuthNUserVerified,
		table: sessionsTable,
	}
	SessionColumnTOTPCheckedAt = Column{
		name:  projection.SessionColumnTOTPCheckedAt,
		table: sessionsTable,
	}
	SessionColumnOTPSMSCheckedAt = Column{
		name:  projection.SessionColumnOTPSMSCheckedAt,
		table: sessionsTable,
	}
	SessionColumnOTPEmailCheckedAt = Column{
		name:  projection.SessionColumnOTPEmailCheckedAt,
		table: sessionsTable,
	}
	SessionColumnMetadata = Column{
		name:  projection.SessionColumnMetadata,
		table: sessionsTable,
	}
	SessionColumnToken = Column{
		name:  projection.SessionColumnTokenID,
		table: sessionsTable,
	}
)

func (q *Queries) SessionByID(ctx context.Context, shouldTriggerBulk bool, id, sessionToken string) (session *Session, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	if shouldTriggerBulk {
		ctx = projection.SessionProjection.Trigger(ctx)
	}

	query, scan := prepareSessionQuery(ctx, q.client)
	stmt, args, err := query.Where(
		sq.Eq{
			SessionColumnID.identifier():         id,
			SessionColumnInstanceID.identifier(): authz.GetInstance(ctx).InstanceID(),
		},
	).ToSql()
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-dn9JW", "Errors.Query.SQLStatement")
	}

	var tokenID string
	err = q.client.QueryRowContext(ctx, func(row *sql.Row) error {
		session, tokenID, err = scan(row)
		return err
	}, stmt, args...)
	if err != nil {
		return nil, err
	}
	if sessionToken == "" {
		return session, nil
	}
	if err := q.sessionTokenVerifier(ctx, sessionToken, session.ID, tokenID); err != nil {
		return nil, errors.ThrowPermissionDenied(nil, "QUERY-dsfr3", "Errors.PermissionDenied")
	}
	return session, nil
}

func (q *Queries) SearchSessions(ctx context.Context, queries *SessionsSearchQueries) (sessions *Sessions, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	query, scan := prepareSessionsQuery(ctx, q.client)
	stmt, args, err := queries.toQuery(query).
		Where(sq.Eq{
			SessionColumnInstanceID.identifier(): authz.GetInstance(ctx).InstanceID(),
		}).ToSql()
	if err != nil {
		return nil, errors.ThrowInvalidArgument(err, "QUERY-sn9Jf", "Errors.Query.InvalidRequest")
	}

	err = q.client.QueryContext(ctx, func(rows *sql.Rows) error {
		sessions, err = scan(rows)
		return err
	}, stmt, args...)
	if err != nil {
		return nil, errors.ThrowInternal(err, "QUERY-Sfg42", "Errors.Internal")
	}

	sessions.LatestSequence, err = q.latestSequence(ctx, sessionsTable)
	return sessions, err
}

func NewSessionIDsSearchQuery(ids []string) (SearchQuery, error) {
	list := make([]interface{}, len(ids))
	for i, value := range ids {
		list[i] = value
	}
	return NewListQuery(SessionColumnID, list, ListIn)
}

func NewSessionCreatorSearchQuery(creator string) (SearchQuery, error) {
	return NewTextQuery(SessionColumnCreator, creator, TextEquals)
}

func prepareSessionQuery(ctx context.Context, db prepareDatabase) (sq.SelectBuilder, func(*sql.Row) (*Session, string, error)) {
	return sq.Select(
			SessionColumnID.identifier(),
			SessionColumnCreationDate.identifier(),
			SessionColumnChangeDate.identifier(),
			SessionColumnSequence.identifier(),
			SessionColumnState.identifier(),
			SessionColumnResourceOwner.identifier(),
			SessionColumnCreator.identifier(),
			SessionColumnUserID.identifier(),
			SessionColumnUserCheckedAt.identifier(),
			LoginNameNameCol.identifier(),
			HumanDisplayNameCol.identifier(),
			UserResourceOwnerCol.identifier(),
			SessionColumnPasswordCheckedAt.identifier(),
			SessionColumnIntentCheckedAt.identifier(),
			SessionColumnWebAuthNCheckedAt.identifier(),
			SessionColumnWebAuthNUserVerified.identifier(),
			SessionColumnTOTPCheckedAt.identifier(),
			SessionColumnOTPSMSCheckedAt.identifier(),
			SessionColumnOTPEmailCheckedAt.identifier(),
			SessionColumnMetadata.identifier(),
			SessionColumnToken.identifier(),
		).From(sessionsTable.identifier()).
			LeftJoin(join(LoginNameUserIDCol, SessionColumnUserID)).
			LeftJoin(join(HumanUserIDCol, SessionColumnUserID)).
			LeftJoin(join(UserIDCol, SessionColumnUserID) + db.Timetravel(call.Took(ctx))).
			PlaceholderFormat(sq.Dollar), func(row *sql.Row) (*Session, string, error) {
			session := new(Session)

			var (
				userID              sql.NullString
				userCheckedAt       sql.NullTime
				loginName           sql.NullString
				displayName         sql.NullString
				userResourceOwner   sql.NullString
				passwordCheckedAt   sql.NullTime
				intentCheckedAt     sql.NullTime
				webAuthNCheckedAt   sql.NullTime
				webAuthNUserPresent sql.NullBool
				totpCheckedAt       sql.NullTime
				otpSMSCheckedAt     sql.NullTime
				otpEmailCheckedAt   sql.NullTime
				metadata            database.Map[[]byte]
				token               sql.NullString
			)

			err := row.Scan(
				&session.ID,
				&session.CreationDate,
				&session.ChangeDate,
				&session.Sequence,
				&session.State,
				&session.ResourceOwner,
				&session.Creator,
				&userID,
				&userCheckedAt,
				&loginName,
				&displayName,
				&userResourceOwner,
				&passwordCheckedAt,
				&intentCheckedAt,
				&webAuthNCheckedAt,
				&webAuthNUserPresent,
				&totpCheckedAt,
				&otpSMSCheckedAt,
				&otpEmailCheckedAt,
				&metadata,
				&token,
			)

			if err != nil {
				if errs.Is(err, sql.ErrNoRows) {
					return nil, "", errors.ThrowNotFound(err, "QUERY-SFeaa", "Errors.Session.NotExisting")
				}
				return nil, "", errors.ThrowInternal(err, "QUERY-SAder", "Errors.Internal")
			}

			session.UserFactor.UserID = userID.String
			session.UserFactor.UserCheckedAt = userCheckedAt.Time
			session.UserFactor.LoginName = loginName.String
			session.UserFactor.DisplayName = displayName.String
			session.UserFactor.ResourceOwner = userResourceOwner.String
			session.PasswordFactor.PasswordCheckedAt = passwordCheckedAt.Time
			session.IntentFactor.IntentCheckedAt = intentCheckedAt.Time
			session.WebAuthNFactor.WebAuthNCheckedAt = webAuthNCheckedAt.Time
			session.WebAuthNFactor.UserVerified = webAuthNUserPresent.Bool
			session.TOTPFactor.TOTPCheckedAt = totpCheckedAt.Time
			session.OTPSMSFactor.OTPCheckedAt = otpSMSCheckedAt.Time
			session.OTPEmailFactor.OTPCheckedAt = otpEmailCheckedAt.Time
			session.Metadata = metadata

			return session, token.String, nil
		}
}

func prepareSessionsQuery(ctx context.Context, db prepareDatabase) (sq.SelectBuilder, func(*sql.Rows) (*Sessions, error)) {
	return sq.Select(
			SessionColumnID.identifier(),
			SessionColumnCreationDate.identifier(),
			SessionColumnChangeDate.identifier(),
			SessionColumnSequence.identifier(),
			SessionColumnState.identifier(),
			SessionColumnResourceOwner.identifier(),
			SessionColumnCreator.identifier(),
			SessionColumnUserID.identifier(),
			SessionColumnUserCheckedAt.identifier(),
			LoginNameNameCol.identifier(),
			HumanDisplayNameCol.identifier(),
			UserResourceOwnerCol.identifier(),
			SessionColumnPasswordCheckedAt.identifier(),
			SessionColumnIntentCheckedAt.identifier(),
			SessionColumnWebAuthNCheckedAt.identifier(),
			SessionColumnWebAuthNUserVerified.identifier(),
			SessionColumnTOTPCheckedAt.identifier(),
			SessionColumnOTPSMSCheckedAt.identifier(),
			SessionColumnOTPEmailCheckedAt.identifier(),
			SessionColumnMetadata.identifier(),
			countColumn.identifier(),
		).From(sessionsTable.identifier()).
			LeftJoin(join(LoginNameUserIDCol, SessionColumnUserID)).
			LeftJoin(join(HumanUserIDCol, SessionColumnUserID)).
			LeftJoin(join(UserIDCol, SessionColumnUserID) + db.Timetravel(call.Took(ctx))).
			PlaceholderFormat(sq.Dollar), func(rows *sql.Rows) (*Sessions, error) {
			sessions := &Sessions{Sessions: []*Session{}}

			for rows.Next() {
				session := new(Session)

				var (
					userID              sql.NullString
					userCheckedAt       sql.NullTime
					loginName           sql.NullString
					displayName         sql.NullString
					userResourceOwner   sql.NullString
					passwordCheckedAt   sql.NullTime
					intentCheckedAt     sql.NullTime
					webAuthNCheckedAt   sql.NullTime
					webAuthNUserPresent sql.NullBool
					totpCheckedAt       sql.NullTime
					otpSMSCheckedAt     sql.NullTime
					otpEmailCheckedAt   sql.NullTime
					metadata            database.Map[[]byte]
				)

				err := rows.Scan(
					&session.ID,
					&session.CreationDate,
					&session.ChangeDate,
					&session.Sequence,
					&session.State,
					&session.ResourceOwner,
					&session.Creator,
					&userID,
					&userCheckedAt,
					&loginName,
					&displayName,
					&userResourceOwner,
					&passwordCheckedAt,
					&intentCheckedAt,
					&webAuthNCheckedAt,
					&webAuthNUserPresent,
					&totpCheckedAt,
					&otpSMSCheckedAt,
					&otpEmailCheckedAt,
					&metadata,
					&sessions.Count,
				)

				if err != nil {
					return nil, errors.ThrowInternal(err, "QUERY-SAfeg", "Errors.Internal")
				}
				session.UserFactor.UserID = userID.String
				session.UserFactor.UserCheckedAt = userCheckedAt.Time
				session.UserFactor.LoginName = loginName.String
				session.UserFactor.DisplayName = displayName.String
				session.UserFactor.ResourceOwner = userResourceOwner.String
				session.PasswordFactor.PasswordCheckedAt = passwordCheckedAt.Time
				session.IntentFactor.IntentCheckedAt = intentCheckedAt.Time
				session.WebAuthNFactor.WebAuthNCheckedAt = webAuthNCheckedAt.Time
				session.WebAuthNFactor.UserVerified = webAuthNUserPresent.Bool
				session.TOTPFactor.TOTPCheckedAt = totpCheckedAt.Time
				session.OTPSMSFactor.OTPCheckedAt = otpSMSCheckedAt.Time
				session.OTPEmailFactor.OTPCheckedAt = otpEmailCheckedAt.Time
				session.Metadata = metadata

				sessions.Sessions = append(sessions.Sessions, session)
			}

			return sessions, nil
		}
}
