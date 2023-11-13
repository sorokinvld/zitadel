package model

import (
	"testing"

	es_models "github.com/zitadel/zitadel/internal/eventstore/v1/models"
	"github.com/zitadel/zitadel/internal/project/model"
	es_model "github.com/zitadel/zitadel/internal/project/repository/eventsourcing/model"
	"github.com/zitadel/zitadel/internal/repository/project"
)

func TestProjectAppendEvent(t *testing.T) {
	type args struct {
		event   *es_models.Event
		project *ProjectView
	}
	tests := []struct {
		name   string
		args   args
		result *ProjectView
	}{
		{
			name: "append added project event",
			args: args{
				event:   &es_models.Event{AggregateID: "AggregateID", Seq: 1, Typ: project.ProjectAddedType, ResourceOwner: "GrantedOrgID", Data: mockProjectData(&es_model.Project{Name: "ProjectName"})},
				project: &ProjectView{},
			},
			result: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectName", State: int32(model.ProjectStateActive)},
		},
		{
			name: "append change project event",
			args: args{
				event:   &es_models.Event{AggregateID: "AggregateID", Seq: 1, Typ: project.ProjectChangedType, ResourceOwner: "GrantedOrgID", Data: mockProjectData(&es_model.Project{Name: "ProjectNameChanged"})},
				project: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectName", State: int32(model.ProjectStateActive)},
			},
			result: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectNameChanged", State: int32(model.ProjectStateActive)},
		},
		{
			name: "append project deactivate event",
			args: args{
				event:   &es_models.Event{AggregateID: "AggregateID", Seq: 1, Typ: project.ProjectDeactivatedType, ResourceOwner: "GrantedOrgID"},
				project: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectName", State: int32(model.ProjectStateActive)},
			},
			result: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectName", State: int32(model.ProjectStateInactive)},
		},
		{
			name: "append project reactivate event",
			args: args{
				event:   &es_models.Event{AggregateID: "AggregateID", Seq: 1, Typ: project.ProjectReactivatedType, ResourceOwner: "GrantedOrgID"},
				project: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectName", State: int32(model.ProjectStateInactive)},
			},
			result: &ProjectView{ProjectID: "AggregateID", ResourceOwner: "GrantedOrgID", Name: "ProjectName", State: int32(model.ProjectStateActive)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.project.AppendEvent(tt.args.event)
			if tt.args.project.ProjectID != tt.result.ProjectID {
				t.Errorf("got wrong result projectID: expected: %v, actual: %v ", tt.result.ProjectID, tt.args.project.ProjectID)
			}
			if tt.args.project.ResourceOwner != tt.result.ResourceOwner {
				t.Errorf("got wrong result ResourceOwner: expected: %v, actual: %v ", tt.result.ResourceOwner, tt.args.project.ResourceOwner)
			}
			if tt.args.project.Name != tt.result.Name {
				t.Errorf("got wrong result name: expected: %v, actual: %v ", tt.result.Name, tt.args.project.Name)
			}
			if tt.args.project.State != tt.result.State {
				t.Errorf("got wrong result state: expected: %v, actual: %v ", tt.result.State, tt.args.project.State)
			}
		})
	}
}
