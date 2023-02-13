package idp

import (
	"encoding/json"

	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/repository"
	"github.com/zitadel/zitadel/internal/repository/idpconfig"
)

type OIDCIDPAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	ID           string              `json:"id"`
	Name         string              `json:"name,omitempty"`
	Issuer       string              `json:"issuer,omitempty"`
	ClientID     string              `json:"client_id,omitempty"`
	ClientSecret *crypto.CryptoValue `json:"client_secret,omitempty"`
	Scopes       []string            `json:"scopes,omitempty"`
	Options
}

func NewOIDCIDPAddedEvent(
	base *eventstore.BaseEvent,
	id,
	name,
	issuer,
	clientID string,
	clientSecret *crypto.CryptoValue,
	scopes []string,
	options Options,
) *OIDCIDPAddedEvent {
	return &OIDCIDPAddedEvent{
		BaseEvent:    *base,
		ID:           id,
		Name:         name,
		Issuer:       issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Options:      options,
	}
}

func (e *OIDCIDPAddedEvent) Data() interface{} {
	return e
}

func (e *OIDCIDPAddedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	return []*eventstore.EventUniqueConstraint{idpconfig.NewAddIDPConfigNameUniqueConstraint(e.Name, e.Aggregate().ResourceOwner)}
}

func OIDCIDPAddedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e := &OIDCIDPAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "IDP-Et1dq", "unable to unmarshal event")
	}

	return e, nil
}

type OIDCIDPChangedEvent struct {
	eventstore.BaseEvent `json:"-"`

	oldName string

	ID           string              `json:"id"`
	Name         *string             `json:"name,omitempty"`
	Issuer       *string             `json:"issuer,omitempty"`
	ClientID     *string             `json:"client_id,omitempty"`
	ClientSecret *crypto.CryptoValue `json:"client_secret,omitempty"`
	Scopes       []string            `json:"scopes,omitempty"` //TODO: tristate?
	OptionChanges
}

func NewOIDCIDPChangedEvent(
	base *eventstore.BaseEvent,
	id,
	oldName string,
	changes []OIDCIDPChanges,
) (*OIDCIDPChangedEvent, error) {
	if len(changes) == 0 {
		return nil, errors.ThrowPreconditionFailed(nil, "IDP-BH3dl", "Errors.NoChangesFound")
	}
	changedEvent := &OIDCIDPChangedEvent{
		BaseEvent: *base,
		oldName:   oldName,
		ID:        id,
	}
	for _, change := range changes {
		change(changedEvent)
	}
	return changedEvent, nil
}

type OIDCIDPChanges func(*OIDCIDPChangedEvent)

func ChangeOIDCName(name string) func(*OIDCIDPChangedEvent) {
	return func(e *OIDCIDPChangedEvent) {
		e.Name = &name
	}
}

func ChangeOIDCIssuer(issuer string) func(*OIDCIDPChangedEvent) {
	return func(e *OIDCIDPChangedEvent) {
		e.Issuer = &issuer
	}
}

func ChangeOIDCClientID(clientID string) func(*OIDCIDPChangedEvent) {
	return func(e *OIDCIDPChangedEvent) {
		e.ClientID = &clientID
	}
}

func ChangeOIDCClientSecret(clientSecret *crypto.CryptoValue) func(*OIDCIDPChangedEvent) {
	return func(e *OIDCIDPChangedEvent) {
		e.ClientSecret = clientSecret
	}
}

func ChangeOIDCOptions(options OptionChanges) func(*OIDCIDPChangedEvent) {
	return func(e *OIDCIDPChangedEvent) {
		e.OptionChanges = options
	}
}

func ChangeOIDCScopes(scopes []string) func(*OIDCIDPChangedEvent) {
	return func(e *OIDCIDPChangedEvent) {
		e.Scopes = scopes
	}
}

func (e *OIDCIDPChangedEvent) Data() interface{} {
	return e
}

func (e *OIDCIDPChangedEvent) UniqueConstraints() []*eventstore.EventUniqueConstraint {
	if e.Name == nil || e.oldName == *e.Name { // TODO: nil check should be enough
		return nil
	}
	return []*eventstore.EventUniqueConstraint{
		idpconfig.NewRemoveIDPConfigNameUniqueConstraint(e.oldName, e.Aggregate().ResourceOwner),
		idpconfig.NewAddIDPConfigNameUniqueConstraint(*e.Name, e.Aggregate().ResourceOwner),
	}
}

func OIDCIDPChangedEventMapper(event *repository.Event) (eventstore.Event, error) {
	e := &OIDCIDPChangedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := json.Unmarshal(event.Data, e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "IDP-D3gjzh", "unable to unmarshal event")
	}

	return e, nil
}