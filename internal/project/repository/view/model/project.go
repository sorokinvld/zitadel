package model

import (
	"encoding/json"
	"time"

	"github.com/zitadel/logging"

	"github.com/zitadel/zitadel/internal/domain"
	caos_errs "github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore/v1/models"
	"github.com/zitadel/zitadel/internal/project/model"
	"github.com/zitadel/zitadel/internal/repository/project"
)

const (
	ProjectKeyProjectID     = "project_id"
	ProjectKeyResourceOwner = "resource_owner"
	ProjectKeyName          = "project_name"
)

type ProjectView struct {
	ProjectID              string                        `json:"-" gorm:"column:project_id;primary_key"`
	Name                   string                        `json:"name" gorm:"column:project_name"`
	CreationDate           time.Time                     `json:"-" gorm:"column:creation_date"`
	ChangeDate             time.Time                     `json:"-" gorm:"column:change_date"`
	State                  int32                         `json:"-" gorm:"column:project_state"`
	ResourceOwner          string                        `json:"-" gorm:"column:resource_owner"`
	ProjectRoleAssertion   bool                          `json:"projectRoleAssertion" gorm:"column:project_role_assertion"`
	ProjectRoleCheck       bool                          `json:"projectRoleCheck" gorm:"column:project_role_check"`
	HasProjectCheck        bool                          `json:"hasProjectCheck" gorm:"column:has_project_check"`
	PrivateLabelingSetting domain.PrivateLabelingSetting `json:"privateLabelingSetting" gorm:"column:private_labeling_setting"`
	Sequence               uint64                        `json:"-" gorm:"column:sequence"`
}

func (p *ProjectView) AppendEvent(event *models.Event) (err error) {
	p.ChangeDate = event.CreationDate
	p.Sequence = event.Seq
	switch event.Type() {
	case project.ProjectAddedType:
		p.State = int32(model.ProjectStateActive)
		p.CreationDate = event.CreationDate
		p.setRootData(event)
		err = p.setData(event)
	case project.ProjectChangedType:
		err = p.setData(event)
	case project.ProjectDeactivatedType:
		p.State = int32(model.ProjectStateInactive)
	case project.ProjectReactivatedType:
		p.State = int32(model.ProjectStateActive)
	case project.ProjectRemovedType:
		p.State = int32(model.ProjectStateRemoved)
	}
	return err
}

func (p *ProjectView) setRootData(event *models.Event) {
	p.ProjectID = event.AggregateID
	p.ResourceOwner = event.ResourceOwner
}

func (p *ProjectView) setData(event *models.Event) error {
	if err := json.Unmarshal(event.Data, p); err != nil {
		logging.Log("EVEN-dlo92").WithError(err).Error("could not unmarshal event data")
		return err
	}
	return nil
}

func (p *ProjectView) setProjectData(event *models.Event) error {
	project := new(ProjectView)
	return project.SetData(event)
}

func (p *ProjectView) SetData(event *models.Event) error {
	if err := json.Unmarshal(event.Data, p); err != nil {
		logging.Log("EVEN-sk9Sj").WithError(err).Error("could not unmarshal event data")
		return caos_errs.ThrowInternal(err, "MODEL-s9ols", "Could not unmarshal data")
	}
	return nil
}
