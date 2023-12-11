package model

import (
	"time"

	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/zerrors"
)

type ProjectGrantMemberView struct {
	UserID             string
	GrantID            string
	ProjectID          string
	UserName           string
	Email              string
	FirstName          string
	LastName           string
	DisplayName        string
	PreferredLoginName string
	AvatarURL          string
	UserResourceOwner  string
	Roles              []string
	CreationDate       time.Time
	ChangeDate         time.Time
	Sequence           uint64
}

type ProjectGrantMemberSearchRequest struct {
	Offset        uint64
	Limit         uint64
	SortingColumn ProjectGrantMemberSearchKey
	Asc           bool
	Queries       []*ProjectGrantMemberSearchQuery
}

type ProjectGrantMemberSearchKey int32

const (
	ProjectGrantMemberSearchKeyUnspecified ProjectGrantMemberSearchKey = iota
	ProjectGrantMemberSearchKeyUserName
	ProjectGrantMemberSearchKeyEmail
	ProjectGrantMemberSearchKeyFirstName
	ProjectGrantMemberSearchKeyLastName
	ProjectGrantMemberSearchKeyGrantID
	ProjectGrantMemberSearchKeyUserID
	ProjectGrantMemberSearchKeyProjectID
)

type ProjectGrantMemberSearchQuery struct {
	Key    ProjectGrantMemberSearchKey
	Method domain.SearchMethod
	Value  interface{}
}

type ProjectGrantMemberSearchResponse struct {
	Offset      uint64
	Limit       uint64
	TotalResult uint64
	Result      []*ProjectGrantMemberView
	Sequence    uint64
	Timestamp   time.Time
}

func (r *ProjectGrantMemberSearchRequest) EnsureLimit(limit uint64) error {
	if r.Limit > limit {
		return zerrors.ThrowInvalidArgument(nil, "SEARCH-ZT8df", "Errors.Limit.ExceedsDefault")
	}
	if r.Limit == 0 {
		r.Limit = limit
	}
	return nil
}
