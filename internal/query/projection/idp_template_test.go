package projection

import (
	"testing"

	"github.com/zitadel/zitadel/internal/database"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/repository"
	"github.com/zitadel/zitadel/internal/repository/instance"
	"github.com/zitadel/zitadel/internal/repository/org"
)

func TestIDPTemplateProjection_reducesOIDC(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceOIDCIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.OIDCIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.OIDCIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOIDCIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeOIDC,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_oidc (idp_id, instance_id, issuer, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5, $6)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"issuer",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceOIDCIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OIDCIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.OIDCIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOIDCIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeOIDC,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_oidc (idp_id, instance_id, issuer, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5, $6)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"issuer",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceOIDCIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.OIDCIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer"
}`),
				), instance.OIDCIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOIDCIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_oidc SET issuer = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"issuer",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceOIDCIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.OIDCIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.OIDCIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOIDCIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8) AND (instance_id = $9)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_oidc SET (client_id, client_secret, issuer, scopes) = ($1, $2, $3, $4) WHERE (idp_id = $5) AND (instance_id = $6)",
							expectedArgs: []interface{}{
								"client_id",
								anyArg{},
								"issuer",
								database.StringArray{"profile"},
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesJWT(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceJWTIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.JWTIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer",
	"jwtEndpoint": "jwt",
	"keysEndpoint": "keys",
	"headerName": "header",
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.JWTIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceJWTIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeJWT,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_jwt (idp_id, instance_id, issuer, jwt_endpoint, keys_endpoint, header_name) VALUES ($1, $2, $3, $4, $5, $6)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"issuer",
								"jwt",
								"keys",
								"header",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceJWTIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.JWTIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer",
	"jwtEndpoint": "jwt",
	"keysEndpoint": "keys",
	"headerName": "header",
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.JWTIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceJWTIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeJWT,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_jwt (idp_id, instance_id, issuer, jwt_endpoint, keys_endpoint, header_name) VALUES ($1, $2, $3, $4, $5, $6)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"issuer",
								"jwt",
								"keys",
								"header",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceJWTIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.JWTIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer"
}`),
				), instance.JWTIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceJWTIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_jwt SET issuer = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"issuer",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceJWTIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.JWTIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"issuer": "issuer",
	"jwtEndpoint": "jwt",
	"keysEndpoint": "keys",
	"headerName": "header",
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.JWTIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceJWTIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8) AND (instance_id = $9)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_jwt SET (jwt_endpoint, keys_endpoint, header_name, issuer) = ($1, $2, $3, $4) WHERE (idp_id = $5) AND (instance_id = $6)",
							expectedArgs: []interface{}{
								"jwt",
								"keys",
								"header",
								"issuer",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesGoogle(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceGoogleIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GoogleIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"clientID": "client_id",
	"clientSecret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.GoogleIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGoogleIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeGoogle,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_google (idp_id, instance_id, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceGoogleIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.GoogleIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"clientID": "client_id",
	"clientSecret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.GoogleIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGoogleIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeGoogle,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_google (idp_id, instance_id, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceGoogleIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GoogleIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"isCreationAllowed": true,
	"clientID": "id"
}`),
				), instance.GoogleIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGoogleIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_google SET client_id = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"id",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceGoogleIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GoogleIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"clientID": "client_id",
	"clientSecret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.GoogleIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGoogleIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6) WHERE (id = $7) AND (instance_id = $8)",
							expectedArgs: []interface{}{
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_google SET (client_id, client_secret, scopes) = ($1, $2, $3) WHERE (idp_id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesOAuth(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceOAuthIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.OAuthIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"authorizationEndpoint": "auth",
	"tokenEndpoint": "token",
 	"userEndpoint": "user",
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.OAuthIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOAuthIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeOAuth,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_oauth (idp_id, instance_id, client_id, client_secret, authorization_endpoint, token_endpoint, user_endpoint, scopes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								"auth",
								"token",
								"user",
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceOAuthIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OAuthIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"authorizationEndpoint": "auth",
	"tokenEndpoint": "token",
 	"userEndpoint": "user",
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.OAuthIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOAuthIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeOAuth,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_oauth (idp_id, instance_id, client_id, client_secret, authorization_endpoint, token_endpoint, user_endpoint, scopes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								"auth",
								"token",
								"user",
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceOAuthIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.OAuthIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"isCreationAllowed": true,
	"client_id": "id"
}`),
				), instance.OAuthIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOAuthIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_oauth SET client_id = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"id",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceOAuthIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.OAuthIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"authorizationEndpoint": "auth",
	"tokenEndpoint": "token",
 	"userEndpoint": "user",
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.OAuthIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceOAuthIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8) AND (instance_id = $9)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_oauth SET (client_id, client_secret, authorization_endpoint, token_endpoint, user_endpoint, scopes) = ($1, $2, $3, $4, $5, $6) WHERE (idp_id = $7) AND (instance_id = $8)",
							expectedArgs: []interface{}{
								"client_id",
								anyArg{},
								"auth",
								"token",
								"user",
								database.StringArray{"profile"},
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesGitHub(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceGitHubIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GitHubIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.GitHubIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitHubIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeGitHub,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_github (idp_id, instance_id, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceGitHubIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.GitHubIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.GitHubIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitHubIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeGitHub,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_github (idp_id, instance_id, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceGitHubIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GitHubIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"isCreationAllowed": true,
	"client_id": "id"
}`),
				), instance.GitHubIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitHubIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_github SET client_id = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"id",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceGitHubIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GitHubIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.GitHubIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitHubIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6) WHERE (id = $7) AND (instance_id = $8)",
							expectedArgs: []interface{}{
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_github SET (client_id, client_secret, scopes) = ($1, $2, $3) WHERE (idp_id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesGitLab(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceGitLabIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GitLabIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.GitLabIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitLabIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeGitLab,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_gitlab (idp_id, instance_id, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceGitLabIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.GitLabIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.GitLabIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitLabIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeGitLab,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_gitlab (idp_id, instance_id, client_id, client_secret, scopes) VALUES ($1, $2, $3, $4, $5)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceGitLabIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GitLabIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"isCreationAllowed": true,
	"client_id": "id"
}`),
				), instance.GitLabIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitLabIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_gitlab SET client_id = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"id",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceGitLabIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.GitLabIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.GitLabIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceGitLabIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6) WHERE (id = $7) AND (instance_id = $8)",
							expectedArgs: []interface{}{
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_gitlab SET (client_id, client_secret, scopes) = ($1, $2, $3) WHERE (idp_id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesAzureAD(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceAzureADIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.AzureADIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"tenant": "tenant",
	"isEmailVerified": true,
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.AzureADIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceAzureADIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeAzureAD,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_azure (idp_id, instance_id, client_id, client_secret, scopes, tenant, is_email_verified) VALUES ($1, $2, $3, $4, $5, $6, $7)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
								"tenant",
								true,
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceAzureADIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.AzureADIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"tenant": "tenant",
	"isEmailVerified": true,
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.AzureADIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceAzureADIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeAzureAD,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_azure (idp_id, instance_id, client_id, client_secret, scopes, tenant, is_email_verified) VALUES ($1, $2, $3, $4, $5, $6, $7)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
								"tenant",
								true,
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceAzureADIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.AzureADIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"isCreationAllowed": true,
	"client_id": "id"
}`),
				), instance.AzureADIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceAzureADIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_azure SET client_id = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"id",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceAzureADIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.AzureADIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"client_id": "client_id",
	"client_secret": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"tenant": "tenant",
	"isEmailVerified": true,
	"scopes": ["profile"],
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.AzureADIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceAzureADIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6) WHERE (id = $7) AND (instance_id = $8)",
							expectedArgs: []interface{}{
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_azure SET (client_id, client_secret, scopes, tenant, is_email_verified) = ($1, $2, $3, $4, $5) WHERE (idp_id = $6) AND (instance_id = $7)",
							expectedArgs: []interface{}{
								"client_id",
								anyArg{},
								database.StringArray{"profile"},
								"tenant",
								true,
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if _, ok := err.(errors.InvalidArgument); !ok {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesRemove(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{

		{
			name: "instance reduceInstanceRemoved",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.InstanceRemovedEventType),
					instance.AggregateType,
					nil,
				), instance.InstanceRemovedEventMapper),
			},
			reduce: reduceInstanceRemovedHelper(IDPTemplateInstanceIDCol),
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.idp_templates WHERE (instance_id = $1)",
							expectedArgs: []interface{}{
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name:   "org reduceOwnerRemoved",
			reduce: (&idpTemplateProjection{}).reduceOwnerRemoved,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgRemovedEventType),
					org.AggregateType,
					nil,
				), org.OrgRemovedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (change_date, sequence, owner_removed) = ($1, $2, $3) WHERE (instance_id = $4) AND (resource_owner = $5)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								true,
								"instance-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
		{
			name:   "org reduceIDPRemoved",
			reduce: (&idpTemplateProjection{}).reduceIDPRemoved,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.IDPRemovedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id"
}`),
				), org.IDPRemovedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "DELETE FROM projections.idp_templates WHERE (id = $1) AND (instance_id = $2)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if !errors.IsErrorInvalidArgument(err) {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}

func TestIDPTemplateProjection_reducesLDAP(t *testing.T) {
	type args struct {
		event func(t *testing.T) eventstore.Event
	}
	tests := []struct {
		name   string
		args   args
		reduce func(event eventstore.Event) (*handler.Statement, error)
		want   wantReduce
	}{
		{
			name: "instance reduceLDAPIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.LDAPIDPAddedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"host": "host",
	"port": "port",
	"tls": true,
	"baseDN": "base",
	"userObjectClass": "user",
	"userUniqueAttribute": "uid",
	"admin": "admin",
	"password": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"idAttribute": "id",
	"firstNameAttribute": "first",
	"lastNameAttribute": "last",
	"displayNameAttribute": "display",
	"nickNameAttribute": "nickname",
	"preferredUsernameAttribute": "username",
	"emailAttribute": "email",
	"emailVerifiedAttribute": "email_verified",
	"phoneAttribute": "phone",
	"phoneVerifiedAttribute": "phone_verified",
	"preferredLanguageAttribute": "lang",
	"avatarURLAttribute": "avatar",
	"profileAttribute": "profile",
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.LDAPIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceLDAPIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeSystem,
								domain.IDPTypeLDAP,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_ldap (idp_id, instance_id, host, port, tls, base_dn, user_object_class, user_unique_attribute, admin, password, id_attribute, first_name_attribute, last_name_attribute, display_name_attribute, nick_name_attribute, preferred_username_attribute, email_attribute, email_verified, phone_attribute, phone_verified_attribute, preferred_language_attribute, avatar_url_attribute, profile_attribute) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"host",
								"port",
								true,
								"base",
								"user",
								"uid",
								"admin",
								anyArg{},
								"id",
								"first",
								"last",
								"display",
								"nickname",
								"username",
								"email",
								"email_verified",
								"phone",
								"phone_verified",
								"lang",
								"avatar",
								"profile",
							},
						},
					},
				},
			},
		},
		{
			name: "org reduceLDAPIDPAdded",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.LDAPIDPAddedEventType),
					org.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"host": "host",
	"port": "port",
	"tls": true,
	"baseDN": "base",
	"userObjectClass": "user",
	"userUniqueAttribute": "uid",
	"admin": "admin",
	"password": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"idAttribute": "id",
	"firstNameAttribute": "first",
	"lastNameAttribute": "last",
	"displayNameAttribute": "display",
	"nickNameAttribute": "nickname",
	"preferredUsernameAttribute": "username",
	"emailAttribute": "email",
	"emailVerifiedAttribute": "email_verified",
	"phoneAttribute": "phone",
	"phoneVerifiedAttribute": "phone_verified",
	"preferredLanguageAttribute": "lang",
	"avatarURLAttribute": "avatar",
	"profileAttribute": "profile",
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), org.LDAPIDPAddedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceLDAPIDPAdded,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "INSERT INTO projections.idp_templates (id, creation_date, change_date, sequence, resource_owner, instance_id, state, name, owner_type, type, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)",
							expectedArgs: []interface{}{
								"idp-id",
								anyArg{},
								anyArg{},
								uint64(15),
								"ro-id",
								"instance-id",
								domain.IDPStateActive,
								"custom-zitadel-instance",
								domain.IdentityProviderTypeOrg,
								domain.IDPTypeLDAP,
								true,
								true,
								true,
								true,
							},
						},
						{
							expectedStmt: "INSERT INTO projections.idp_templates_ldap (idp_id, instance_id, host, port, tls, base_dn, user_object_class, user_unique_attribute, admin, password, id_attribute, first_name_attribute, last_name_attribute, display_name_attribute, nick_name_attribute, preferred_username_attribute, email_attribute, email_verified, phone_attribute, phone_verified_attribute, preferred_language_attribute, avatar_url_attribute, profile_attribute) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)",
							expectedArgs: []interface{}{
								"idp-id",
								"instance-id",
								"host",
								"port",
								true,
								"base",
								"user",
								"uid",
								"admin",
								anyArg{},
								"id",
								"first",
								"last",
								"display",
								"nickname",
								"username",
								"email",
								"email_verified",
								"phone",
								"phone_verified",
								"lang",
								"avatar",
								"profile",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceLDAPIDPChanged minimal",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.LDAPIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"host": "host"
}`),
				), instance.LDAPIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceLDAPIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, change_date, sequence) = ($1, $2, $3) WHERE (id = $4) AND (instance_id = $5)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_ldap SET host = $1 WHERE (idp_id = $2) AND (instance_id = $3)",
							expectedArgs: []interface{}{
								"host",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name: "instance reduceLDAPIDPChanged",
			args: args{
				event: getEvent(testEvent(
					repository.EventType(instance.LDAPIDPChangedEventType),
					instance.AggregateType,
					[]byte(`{
	"id": "idp-id",
	"name": "custom-zitadel-instance",
	"host": "host",
	"port": "port",
	"tls": true,
	"baseDN": "base",
	"userObjectClass": "user",
	"userUniqueAttribute": "uid",
	"admin": "admin",
	"password": {
        "cryptoType": 0,
        "algorithm": "RSA-265",
        "keyId": "key-id"
    },
	"idAttribute": "id",
	"firstNameAttribute": "first",
	"lastNameAttribute": "last",
	"displayNameAttribute": "display",
	"nickNameAttribute": "nickname",
	"preferredUsernameAttribute": "username",
	"emailAttribute": "email",
	"emailVerifiedAttribute": "email_verified",
	"phoneAttribute": "phone",
	"phoneVerifiedAttribute": "phone_verified",
	"preferredLanguageAttribute": "lang",
	"avatarURLAttribute": "avatar",
	"profileAttribute": "profile",
	"isCreationAllowed": true,
	"isLinkingAllowed": true,
	"isAutoCreation": true,
	"isAutoUpdate": true
}`),
				), instance.LDAPIDPChangedEventMapper),
			},
			reduce: (&idpTemplateProjection{}).reduceLDAPIDPChanged,
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("instance"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (name, is_creation_allowed, is_linking_allowed, is_auto_creation, is_auto_update, change_date, sequence) = ($1, $2, $3, $4, $5, $6, $7) WHERE (id = $8) AND (instance_id = $9)",
							expectedArgs: []interface{}{
								"custom-zitadel-instance",
								true,
								true,
								true,
								true,
								anyArg{},
								uint64(15),
								"idp-id",
								"instance-id",
							},
						},
						{
							expectedStmt: "UPDATE projections.idp_templates_ldap SET (host, port, tls, base_dn, user_object_class, user_unique_attribute, admin, password, id_attribute, first_name_attribute, last_name_attribute, display_name_attribute, nick_name_attribute, preferred_username_attribute, email_attribute, email_verified, phone_attribute, phone_verified_attribute, preferred_language_attribute, avatar_url_attribute, profile_attribute) = ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21) WHERE (idp_id = $22) AND (instance_id = $23)",
							expectedArgs: []interface{}{
								"host",
								"port",
								true,
								"base",
								"user",
								"uid",
								"admin",
								anyArg{},
								"id",
								"first",
								"last",
								"display",
								"nickname",
								"username",
								"email",
								"email_verified",
								"phone",
								"phone_verified",
								"lang",
								"avatar",
								"profile",
								"idp-id",
								"instance-id",
							},
						},
					},
				},
			},
		},
		{
			name:   "org.reduceOwnerRemoved",
			reduce: (&idpTemplateProjection{}).reduceOwnerRemoved,
			args: args{
				event: getEvent(testEvent(
					repository.EventType(org.OrgRemovedEventType),
					org.AggregateType,
					nil,
				), org.OrgRemovedEventMapper),
			},
			want: wantReduce{
				aggregateType:    eventstore.AggregateType("org"),
				sequence:         15,
				previousSequence: 10,
				executer: &testExecuter{
					executions: []execution{
						{
							expectedStmt: "UPDATE projections.idp_templates SET (change_date, sequence, owner_removed) = ($1, $2, $3) WHERE (instance_id = $4) AND (resource_owner = $5)",
							expectedArgs: []interface{}{
								anyArg{},
								uint64(15),
								true,
								"instance-id",
								"agg-id",
							},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := baseEvent(t)
			got, err := tt.reduce(event)
			if !errors.IsErrorInvalidArgument(err) {
				t.Errorf("no wrong event mapping: %v, got: %v", err, got)
			}

			event = tt.args.event(t)
			got, err = tt.reduce(event)
			assertReduce(t, got, err, IDPTemplateTable, tt.want)
		})
	}
}
