package command

import (
	"reflect"
	"time"

	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/repository/idp"
	"github.com/zitadel/zitadel/internal/repository/idpconfig"
)

type OAuthIDPWriteModel struct {
	eventstore.WriteModel

	Name                  string
	ID                    string
	ClientID              string
	ClientSecret          *crypto.CryptoValue
	AuthorizationEndpoint string
	TokenEndpoint         string
	UserEndpoint          string
	Scopes                []string
	IDAttribute           string
	idp.Options

	State domain.IDPState
}

func (wm *OAuthIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.OAuthIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.OAuthIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *OAuthIDPWriteModel) reduceAddedEvent(e *idp.OAuthIDPAddedEvent) {
	wm.Name = e.Name
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.AuthorizationEndpoint = e.AuthorizationEndpoint
	wm.TokenEndpoint = e.TokenEndpoint
	wm.UserEndpoint = e.UserEndpoint
	wm.Scopes = e.Scopes
	wm.IDAttribute = e.IDAttribute
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *OAuthIDPWriteModel) reduceChangedEvent(e *idp.OAuthIDPChangedEvent) {
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.AuthorizationEndpoint != nil {
		wm.AuthorizationEndpoint = *e.AuthorizationEndpoint
	}
	if e.TokenEndpoint != nil {
		wm.TokenEndpoint = *e.TokenEndpoint
	}
	if e.UserEndpoint != nil {
		wm.UserEndpoint = *e.UserEndpoint
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	if e.IDAttribute != nil {
		wm.IDAttribute = *e.IDAttribute
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *OAuthIDPWriteModel) NewChanges(
	name,
	clientID,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	authorizationEndpoint,
	tokenEndpoint,
	userEndpoint,
	idAttribute string,
	scopes []string,
	options idp.Options,
) ([]idp.OAuthIDPChanges, error) {
	changes := make([]idp.OAuthIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeOAuthClientSecret(clientSecret))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeOAuthClientID(clientID))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeOAuthName(name))
	}
	if wm.AuthorizationEndpoint != authorizationEndpoint {
		changes = append(changes, idp.ChangeOAuthAuthorizationEndpoint(authorizationEndpoint))
	}
	if wm.TokenEndpoint != tokenEndpoint {
		changes = append(changes, idp.ChangeOAuthTokenEndpoint(tokenEndpoint))
	}
	if wm.UserEndpoint != userEndpoint {
		changes = append(changes, idp.ChangeOAuthUserEndpoint(userEndpoint))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeOAuthScopes(scopes))
	}
	if wm.IDAttribute != idAttribute {
		changes = append(changes, idp.ChangeOAuthIDAttribute(idAttribute))
	}
	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeOAuthOptions(opts))
	}
	return changes, nil
}

type OIDCIDPWriteModel struct {
	eventstore.WriteModel

	Name             string
	ID               string
	Issuer           string
	ClientID         string
	ClientSecret     *crypto.CryptoValue
	Scopes           []string
	IsIDTokenMapping bool
	idp.Options

	State domain.IDPState
}

func (wm *OIDCIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.OIDCIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.OIDCIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idpconfig.IDPConfigAddedEvent:
			wm.reduceIDPConfigAddedEvent(e)
		case *idpconfig.IDPConfigChangedEvent:
			wm.reduceIDPConfigChangedEvent(e)
		case *idpconfig.OIDCConfigAddedEvent:
			wm.reduceOIDCConfigAddedEvent(e)
		case *idpconfig.OIDCConfigChangedEvent:
			wm.reduceOIDCConfigChangedEvent(e)
		case *idpconfig.IDPConfigRemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *OIDCIDPWriteModel) reduceAddedEvent(e *idp.OIDCIDPAddedEvent) {
	wm.Name = e.Name
	wm.Issuer = e.Issuer
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
	wm.IsIDTokenMapping = e.IsIDTokenMapping
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *OIDCIDPWriteModel) reduceChangedEvent(e *idp.OIDCIDPChangedEvent) {
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.Issuer != nil {
		wm.Issuer = *e.Issuer
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	if e.IsIDTokenMapping != nil {
		wm.IsIDTokenMapping = *e.IsIDTokenMapping
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *OIDCIDPWriteModel) NewChanges(
	name,
	issuer,
	clientID,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	scopes []string,
	idTokenMapping bool,
	options idp.Options,
) ([]idp.OIDCIDPChanges, error) {
	changes := make([]idp.OIDCIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeOIDCClientSecret(clientSecret))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeOIDCClientID(clientID))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeOIDCName(name))
	}
	if wm.Issuer != issuer {
		changes = append(changes, idp.ChangeOIDCIssuer(issuer))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeOIDCScopes(scopes))
	}
	if wm.IsIDTokenMapping != idTokenMapping {
		changes = append(changes, idp.ChangeOIDCIsIDTokenMapping(idTokenMapping))
	}
	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeOIDCOptions(opts))
	}
	return changes, nil
}

// reduceIDPConfigAddedEvent handles old idpConfig events
func (wm *OIDCIDPWriteModel) reduceIDPConfigAddedEvent(e *idpconfig.IDPConfigAddedEvent) {
	wm.Name = e.Name
	wm.Options.IsAutoCreation = e.AutoRegister
	wm.State = domain.IDPStateActive
}

// reduceIDPConfigChangedEvent handles old idpConfig changes
func (wm *OIDCIDPWriteModel) reduceIDPConfigChangedEvent(e *idpconfig.IDPConfigChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.AutoRegister != nil {
		wm.Options.IsAutoCreation = *e.AutoRegister
	}
}

// reduceOIDCConfigAddedEvent handles old OIDC idpConfig events
func (wm *OIDCIDPWriteModel) reduceOIDCConfigAddedEvent(e *idpconfig.OIDCConfigAddedEvent) {
	wm.Issuer = e.Issuer
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
}

// reduceOIDCConfigChangedEvent handles old OIDC idpConfig changes
func (wm *OIDCIDPWriteModel) reduceOIDCConfigChangedEvent(e *idpconfig.OIDCConfigChangedEvent) {
	if e.Issuer != nil {
		wm.Issuer = *e.Issuer
	}
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
}

type JWTIDPWriteModel struct {
	eventstore.WriteModel

	ID           string
	Name         string
	Issuer       string
	JWTEndpoint  string
	KeysEndpoint string
	HeaderName   string
	idp.Options

	State domain.IDPState
}

func (wm *JWTIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.JWTIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.JWTIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idpconfig.IDPConfigAddedEvent:
			wm.reduceIDPConfigAddedEvent(e)
		case *idpconfig.IDPConfigChangedEvent:
			wm.reduceIDPConfigChangedEvent(e)
		case *idpconfig.JWTConfigAddedEvent:
			wm.reduceJWTConfigAddedEvent(e)
		case *idpconfig.JWTConfigChangedEvent:
			wm.reduceJWTConfigChangedEvent(e)
		case *idpconfig.IDPConfigRemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *JWTIDPWriteModel) reduceAddedEvent(e *idp.JWTIDPAddedEvent) {
	wm.Name = e.Name
	wm.Issuer = e.Issuer
	wm.JWTEndpoint = e.JWTEndpoint
	wm.KeysEndpoint = e.KeysEndpoint
	wm.HeaderName = e.HeaderName
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *JWTIDPWriteModel) reduceChangedEvent(e *idp.JWTIDPChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.Issuer != nil {
		wm.Issuer = *e.Issuer
	}
	if e.JWTEndpoint != nil {
		wm.JWTEndpoint = *e.JWTEndpoint
	}
	if e.KeysEndpoint != nil {
		wm.KeysEndpoint = *e.KeysEndpoint
	}
	if e.HeaderName != nil {
		wm.HeaderName = *e.HeaderName
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *JWTIDPWriteModel) NewChanges(
	name,
	issuer,
	jwtEndpoint,
	keysEndpoint,
	headerName string,
	options idp.Options,
) ([]idp.JWTIDPChanges, error) {
	changes := make([]idp.JWTIDPChanges, 0)
	if wm.Name != name {
		changes = append(changes, idp.ChangeJWTName(name))
	}
	if wm.Issuer != issuer {
		changes = append(changes, idp.ChangeJWTIssuer(issuer))
	}
	if wm.JWTEndpoint != jwtEndpoint {
		changes = append(changes, idp.ChangeJWTEndpoint(jwtEndpoint))
	}
	if wm.KeysEndpoint != keysEndpoint {
		changes = append(changes, idp.ChangeJWTKeysEndpoint(keysEndpoint))
	}
	if wm.HeaderName != headerName {
		changes = append(changes, idp.ChangeJWTHeaderName(headerName))
	}
	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeJWTOptions(opts))
	}
	return changes, nil
}

// reduceIDPConfigAddedEvent handles old idpConfig events
func (wm *JWTIDPWriteModel) reduceIDPConfigAddedEvent(e *idpconfig.IDPConfigAddedEvent) {
	wm.Name = e.Name
	wm.Options.IsAutoCreation = e.AutoRegister
	wm.State = domain.IDPStateActive
}

// reduceIDPConfigChangedEvent handles old idpConfig changes
func (wm *JWTIDPWriteModel) reduceIDPConfigChangedEvent(e *idpconfig.IDPConfigChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.AutoRegister != nil {
		wm.Options.IsAutoCreation = *e.AutoRegister
	}
}

// reduceJWTConfigAddedEvent handles old JWT idpConfig events
func (wm *JWTIDPWriteModel) reduceJWTConfigAddedEvent(e *idpconfig.JWTConfigAddedEvent) {
	wm.Issuer = e.Issuer
	wm.JWTEndpoint = e.JWTEndpoint
	wm.KeysEndpoint = e.KeysEndpoint
	wm.HeaderName = e.HeaderName
}

// reduceJWTConfigChangedEvent handles old JWT idpConfig changes
func (wm *JWTIDPWriteModel) reduceJWTConfigChangedEvent(e *idpconfig.JWTConfigChangedEvent) {
	if e.Issuer != nil {
		wm.Issuer = *e.Issuer
	}
	if e.JWTEndpoint != nil {
		wm.JWTEndpoint = *e.JWTEndpoint
	}
	if e.KeysEndpoint != nil {
		wm.KeysEndpoint = *e.KeysEndpoint
	}
	if e.HeaderName != nil {
		wm.HeaderName = *e.HeaderName
	}
}

type AzureADIDPWriteModel struct {
	eventstore.WriteModel

	ID              string
	Name            string
	ClientID        string
	ClientSecret    *crypto.CryptoValue
	Scopes          []string
	Tenant          string
	IsEmailVerified bool
	idp.Options

	State domain.IDPState
}

func (wm *AzureADIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.AzureADIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.AzureADIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *AzureADIDPWriteModel) reduceAddedEvent(e *idp.AzureADIDPAddedEvent) {
	wm.Name = e.Name
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
	wm.Tenant = e.Tenant
	wm.IsEmailVerified = e.IsEmailVerified
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *AzureADIDPWriteModel) reduceChangedEvent(e *idp.AzureADIDPChangedEvent) {
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	if e.Tenant != nil {
		wm.Tenant = *e.Tenant
	}
	if e.IsEmailVerified != nil {
		wm.IsEmailVerified = *e.IsEmailVerified
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *AzureADIDPWriteModel) NewChanges(
	name string,
	clientID string,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	scopes []string,
	tenant string,
	isEmailVerified bool,
	options idp.Options,
) ([]idp.AzureADIDPChanges, error) {
	changes := make([]idp.AzureADIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeAzureADClientSecret(clientSecret))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeAzureADName(name))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeAzureADClientID(clientID))
	}
	if wm.Tenant != tenant {
		changes = append(changes, idp.ChangeAzureADTenant(tenant))
	}
	if wm.IsEmailVerified != isEmailVerified {
		changes = append(changes, idp.ChangeAzureADIsEmailVerified(isEmailVerified))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeAzureADScopes(scopes))
	}

	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeAzureADOptions(opts))
	}
	return changes, nil
}

type GitHubIDPWriteModel struct {
	eventstore.WriteModel

	ID           string
	Name         string
	ClientID     string
	ClientSecret *crypto.CryptoValue
	Scopes       []string
	idp.Options

	State domain.IDPState
}

func (wm *GitHubIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.GitHubIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.GitHubIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *GitHubIDPWriteModel) reduceAddedEvent(e *idp.GitHubIDPAddedEvent) {
	wm.Name = e.Name
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *GitHubIDPWriteModel) reduceChangedEvent(e *idp.GitHubIDPChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *GitHubIDPWriteModel) NewChanges(
	name,
	clientID,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	scopes []string,
	options idp.Options,
) ([]idp.GitHubIDPChanges, error) {
	changes := make([]idp.GitHubIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeGitHubClientSecret(clientSecret))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeGitHubName(name))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeGitHubClientID(clientID))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeGitHubScopes(scopes))
	}

	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeGitHubOptions(opts))
	}
	return changes, nil
}

type GitHubEnterpriseIDPWriteModel struct {
	eventstore.WriteModel

	ID                    string
	Name                  string
	ClientID              string
	ClientSecret          *crypto.CryptoValue
	AuthorizationEndpoint string
	TokenEndpoint         string
	UserEndpoint          string
	Scopes                []string
	idp.Options

	State domain.IDPState
}

func (wm *GitHubEnterpriseIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.GitHubEnterpriseIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.GitHubEnterpriseIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *GitHubEnterpriseIDPWriteModel) reduceAddedEvent(e *idp.GitHubEnterpriseIDPAddedEvent) {
	wm.Name = e.Name
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.AuthorizationEndpoint = e.AuthorizationEndpoint
	wm.TokenEndpoint = e.TokenEndpoint
	wm.UserEndpoint = e.UserEndpoint
	wm.Scopes = e.Scopes
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *GitHubEnterpriseIDPWriteModel) reduceChangedEvent(e *idp.GitHubEnterpriseIDPChangedEvent) {
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.AuthorizationEndpoint != nil {
		wm.AuthorizationEndpoint = *e.AuthorizationEndpoint
	}
	if e.TokenEndpoint != nil {
		wm.TokenEndpoint = *e.TokenEndpoint
	}
	if e.UserEndpoint != nil {
		wm.UserEndpoint = *e.UserEndpoint
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *GitHubEnterpriseIDPWriteModel) NewChanges(
	name,
	clientID string,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	authorizationEndpoint,
	tokenEndpoint,
	userEndpoint string,
	scopes []string,
	options idp.Options,
) ([]idp.GitHubEnterpriseIDPChanges, error) {
	changes := make([]idp.GitHubEnterpriseIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeGitHubEnterpriseClientSecret(clientSecret))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeGitHubEnterpriseClientID(clientID))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeGitHubEnterpriseName(name))
	}
	if wm.AuthorizationEndpoint != authorizationEndpoint {
		changes = append(changes, idp.ChangeGitHubEnterpriseAuthorizationEndpoint(authorizationEndpoint))
	}
	if wm.TokenEndpoint != tokenEndpoint {
		changes = append(changes, idp.ChangeGitHubEnterpriseTokenEndpoint(tokenEndpoint))
	}
	if wm.UserEndpoint != userEndpoint {
		changes = append(changes, idp.ChangeGitHubEnterpriseUserEndpoint(userEndpoint))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeGitHubEnterpriseScopes(scopes))
	}
	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeGitHubEnterpriseOptions(opts))
	}
	return changes, nil
}

type GitLabIDPWriteModel struct {
	eventstore.WriteModel

	ID           string
	Name         string
	ClientID     string
	ClientSecret *crypto.CryptoValue
	Scopes       []string
	idp.Options

	State domain.IDPState
}

func (wm *GitLabIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.GitLabIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.GitLabIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *GitLabIDPWriteModel) reduceAddedEvent(e *idp.GitLabIDPAddedEvent) {
	wm.Name = e.Name
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *GitLabIDPWriteModel) reduceChangedEvent(e *idp.GitLabIDPChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *GitLabIDPWriteModel) NewChanges(
	name,
	clientID,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	scopes []string,
	options idp.Options,
) ([]idp.GitLabIDPChanges, error) {
	changes := make([]idp.GitLabIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeGitLabClientSecret(clientSecret))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeGitLabName(name))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeGitLabClientID(clientID))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeGitLabScopes(scopes))
	}

	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeGitLabOptions(opts))
	}
	return changes, nil
}

type GitLabSelfHostedIDPWriteModel struct {
	eventstore.WriteModel

	ID           string
	Name         string
	Issuer       string
	ClientID     string
	ClientSecret *crypto.CryptoValue
	Scopes       []string
	idp.Options

	State domain.IDPState
}

func (wm *GitLabSelfHostedIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.GitLabSelfHostedIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.GitLabSelfHostedIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *GitLabSelfHostedIDPWriteModel) reduceAddedEvent(e *idp.GitLabSelfHostedIDPAddedEvent) {
	wm.Name = e.Name
	wm.Issuer = e.Issuer
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *GitLabSelfHostedIDPWriteModel) reduceChangedEvent(e *idp.GitLabSelfHostedIDPChangedEvent) {
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.Issuer != nil {
		wm.Issuer = *e.Issuer
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *GitLabSelfHostedIDPWriteModel) NewChanges(
	name string,
	issuer string,
	clientID string,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	scopes []string,
	options idp.Options,
) ([]idp.GitLabSelfHostedIDPChanges, error) {
	changes := make([]idp.GitLabSelfHostedIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeGitLabSelfHostedClientSecret(clientSecret))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeGitLabSelfHostedClientID(clientID))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeGitLabSelfHostedName(name))
	}
	if wm.Issuer != issuer {
		changes = append(changes, idp.ChangeGitLabSelfHostedIssuer(issuer))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeGitLabSelfHostedScopes(scopes))
	}
	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeGitLabSelfHostedOptions(opts))
	}
	return changes, nil
}

type GoogleIDPWriteModel struct {
	eventstore.WriteModel

	ID           string
	Name         string
	ClientID     string
	ClientSecret *crypto.CryptoValue
	Scopes       []string
	idp.Options

	State domain.IDPState
}

func (wm *GoogleIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.GoogleIDPAddedEvent:
			wm.reduceAddedEvent(e)
		case *idp.GoogleIDPChangedEvent:
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *GoogleIDPWriteModel) reduceAddedEvent(e *idp.GoogleIDPAddedEvent) {
	wm.Name = e.Name
	wm.ClientID = e.ClientID
	wm.ClientSecret = e.ClientSecret
	wm.Scopes = e.Scopes
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *GoogleIDPWriteModel) reduceChangedEvent(e *idp.GoogleIDPChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.ClientID != nil {
		wm.ClientID = *e.ClientID
	}
	if e.ClientSecret != nil {
		wm.ClientSecret = e.ClientSecret
	}
	if e.Scopes != nil {
		wm.Scopes = e.Scopes
	}
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *GoogleIDPWriteModel) NewChanges(
	name string,
	clientID string,
	clientSecretString string,
	secretCrypto crypto.Crypto,
	scopes []string,
	options idp.Options,
) ([]idp.GoogleIDPChanges, error) {
	changes := make([]idp.GoogleIDPChanges, 0)
	var clientSecret *crypto.CryptoValue
	var err error
	if clientSecretString != "" {
		clientSecret, err = crypto.Crypt([]byte(clientSecretString), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeGoogleClientSecret(clientSecret))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeGoogleName(name))
	}
	if wm.ClientID != clientID {
		changes = append(changes, idp.ChangeGoogleClientID(clientID))
	}
	if !reflect.DeepEqual(wm.Scopes, scopes) {
		changes = append(changes, idp.ChangeGoogleScopes(scopes))
	}

	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeGoogleOptions(opts))
	}
	return changes, nil
}

type LDAPIDPWriteModel struct {
	eventstore.WriteModel

	ID                string
	Name              string
	Servers           []string
	StartTLS          bool
	BaseDN            string
	BindDN            string
	BindPassword      *crypto.CryptoValue
	UserBase          string
	UserObjectClasses []string
	UserFilters       []string
	Timeout           time.Duration
	idp.LDAPAttributes
	idp.Options

	State domain.IDPState
}

func (wm *LDAPIDPWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.LDAPIDPAddedEvent:
			if wm.ID != e.ID {
				continue
			}
			wm.reduceAddedEvent(e)
		case *idp.LDAPIDPChangedEvent:
			if wm.ID != e.ID {
				continue
			}
			wm.reduceChangedEvent(e)
		case *idp.RemovedEvent:
			if wm.ID != e.ID {
				continue
			}
			wm.State = domain.IDPStateRemoved
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *LDAPIDPWriteModel) reduceAddedEvent(e *idp.LDAPIDPAddedEvent) {
	wm.Name = e.Name
	wm.Servers = e.Servers
	wm.StartTLS = e.StartTLS
	wm.BaseDN = e.BaseDN
	wm.BindDN = e.BindDN
	wm.BindPassword = e.BindPassword
	wm.UserBase = e.UserBase
	wm.UserObjectClasses = e.UserObjectClasses
	wm.UserFilters = e.UserFilters
	wm.Timeout = e.Timeout
	wm.LDAPAttributes = e.LDAPAttributes
	wm.Options = e.Options
	wm.State = domain.IDPStateActive
}

func (wm *LDAPIDPWriteModel) reduceChangedEvent(e *idp.LDAPIDPChangedEvent) {
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.Name != nil {
		wm.Name = *e.Name
	}
	if e.Servers != nil {
		wm.Servers = e.Servers
	}
	if e.StartTLS != nil {
		wm.StartTLS = *e.StartTLS
	}
	if e.BaseDN != nil {
		wm.BaseDN = *e.BaseDN
	}
	if e.BindDN != nil {
		wm.BindDN = *e.BindDN
	}
	if e.BindPassword != nil {
		wm.BindPassword = e.BindPassword
	}
	if e.UserBase != nil {
		wm.UserBase = *e.UserBase
	}
	if e.UserObjectClasses != nil {
		wm.UserObjectClasses = e.UserObjectClasses
	}
	if e.UserFilters != nil {
		wm.UserFilters = e.UserFilters
	}
	if e.Timeout != nil {
		wm.Timeout = *e.Timeout
	}
	wm.LDAPAttributes.ReduceChanges(e.LDAPAttributeChanges)
	wm.Options.ReduceChanges(e.OptionChanges)
}

func (wm *LDAPIDPWriteModel) NewChanges(
	name string,
	servers []string,
	startTLS bool,
	baseDN string,
	bindDN string,
	bindPassword string,
	userBase string,
	userObjectClasses []string,
	userFilters []string,
	timeout time.Duration,
	secretCrypto crypto.Crypto,
	attributes idp.LDAPAttributes,
	options idp.Options,
) ([]idp.LDAPIDPChanges, error) {
	changes := make([]idp.LDAPIDPChanges, 0)
	var cryptedPassword *crypto.CryptoValue
	var err error
	if bindPassword != "" {
		cryptedPassword, err = crypto.Crypt([]byte(bindPassword), secretCrypto)
		if err != nil {
			return nil, err
		}
		changes = append(changes, idp.ChangeLDAPBindPassword(cryptedPassword))
	}
	if wm.Name != name {
		changes = append(changes, idp.ChangeLDAPName(name))
	}
	if !reflect.DeepEqual(wm.Servers, servers) {
		changes = append(changes, idp.ChangeLDAPServers(servers))
	}
	if wm.StartTLS != startTLS {
		changes = append(changes, idp.ChangeLDAPStartTLS(startTLS))
	}
	if wm.BaseDN != baseDN {
		changes = append(changes, idp.ChangeLDAPBaseDN(baseDN))
	}
	if wm.BindDN != bindDN {
		changes = append(changes, idp.ChangeLDAPBindDN(bindDN))
	}
	if wm.UserBase != userBase {
		changes = append(changes, idp.ChangeLDAPUserBase(userBase))
	}
	if !reflect.DeepEqual(wm.UserObjectClasses, userObjectClasses) {
		changes = append(changes, idp.ChangeLDAPUserObjectClasses(userObjectClasses))
	}
	if !reflect.DeepEqual(wm.UserFilters, userFilters) {
		changes = append(changes, idp.ChangeLDAPUserFilters(userFilters))
	}
	if wm.Timeout != timeout {
		changes = append(changes, idp.ChangeLDAPTimeout(timeout))
	}
	attrs := wm.LDAPAttributes.Changes(attributes)
	if !attrs.IsZero() {
		changes = append(changes, idp.ChangeLDAPAttributes(attrs))
	}
	opts := wm.Options.Changes(options)
	if !opts.IsZero() {
		changes = append(changes, idp.ChangeLDAPOptions(opts))
	}
	return changes, nil
}

type IDPRemoveWriteModel struct {
	eventstore.WriteModel

	ID    string
	State domain.IDPState
}

func (wm *IDPRemoveWriteModel) Reduce() error {
	for _, event := range wm.Events {
		switch e := event.(type) {
		case *idp.OAuthIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.OIDCIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.JWTIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.AzureADIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.GitHubIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.GitHubEnterpriseIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.GitLabIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.GitLabSelfHostedIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.GoogleIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.LDAPIDPAddedEvent:
			wm.reduceAdded(e.ID)
		case *idp.RemovedEvent:
			wm.reduceRemoved(e.ID)
		case *idpconfig.IDPConfigAddedEvent:
			wm.reduceAdded(e.ConfigID)
		case *idpconfig.IDPConfigRemovedEvent:
			wm.reduceRemoved(e.ConfigID)
		}
	}
	return wm.WriteModel.Reduce()
}

func (wm *IDPRemoveWriteModel) reduceAdded(id string) {
	if wm.ID != id {
		return
	}
	wm.State = domain.IDPStateActive
}

func (wm *IDPRemoveWriteModel) reduceRemoved(id string) {
	if wm.ID != id {
		return
	}
	wm.State = domain.IDPStateRemoved
}
