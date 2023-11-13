package idp

import (
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
)

type Options struct {
	IsCreationAllowed bool `json:"isCreationAllowed,omitempty"`
	IsLinkingAllowed  bool `json:"isLinkingAllowed,omitempty"`
	IsAutoCreation    bool `json:"isAutoCreation,omitempty"`
	IsAutoUpdate      bool `json:"isAutoUpdate,omitempty"`
}

type OptionChanges struct {
	IsCreationAllowed *bool `json:"isCreationAllowed,omitempty"`
	IsLinkingAllowed  *bool `json:"isLinkingAllowed,omitempty"`
	IsAutoCreation    *bool `json:"isAutoCreation,omitempty"`
	IsAutoUpdate      *bool `json:"isAutoUpdate,omitempty"`
}

func (o *Options) Changes(options Options) OptionChanges {
	opts := OptionChanges{}
	if o.IsCreationAllowed != options.IsCreationAllowed {
		opts.IsCreationAllowed = &options.IsCreationAllowed
	}
	if o.IsLinkingAllowed != options.IsLinkingAllowed {
		opts.IsLinkingAllowed = &options.IsLinkingAllowed
	}
	if o.IsAutoCreation != options.IsAutoCreation {
		opts.IsAutoCreation = &options.IsAutoCreation
	}
	if o.IsAutoUpdate != options.IsAutoUpdate {
		opts.IsAutoUpdate = &options.IsAutoUpdate
	}
	return opts
}

func (o *Options) ReduceChanges(changes OptionChanges) {
	if changes.IsCreationAllowed != nil {
		o.IsCreationAllowed = *changes.IsCreationAllowed
	}
	if changes.IsLinkingAllowed != nil {
		o.IsLinkingAllowed = *changes.IsLinkingAllowed
	}
	if changes.IsAutoCreation != nil {
		o.IsAutoCreation = *changes.IsAutoCreation
	}
	if changes.IsAutoUpdate != nil {
		o.IsAutoUpdate = *changes.IsAutoUpdate
	}
}

func (o *OptionChanges) IsZero() bool {
	return o.IsCreationAllowed == nil && o.IsLinkingAllowed == nil && o.IsAutoCreation == nil && o.IsAutoUpdate == nil
}

type RemovedEvent struct {
	eventstore.BaseEvent `json:"-"`

	ID string `json:"id"`
}

func NewRemovedEvent(
	base *eventstore.BaseEvent,
	id string,
) *RemovedEvent {
	return &RemovedEvent{
		BaseEvent: *base,
		ID:        id,
	}
}

func (e *RemovedEvent) Payload() interface{} {
	return e
}

func (e *RemovedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return nil
}

func RemovedEventMapper(event eventstore.Event) (eventstore.Event, error) {
	e := &RemovedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}

	err := event.Unmarshal(e)
	if err != nil {
		return nil, errors.ThrowInternal(err, "IDP-plSD2", "unable to unmarshal event")
	}

	return e, nil
}
