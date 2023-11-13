package settings

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/api/grpc/object/v2"
	"github.com/zitadel/zitadel/internal/api/grpc/text"
	"github.com/zitadel/zitadel/internal/query"
	object_pb "github.com/zitadel/zitadel/pkg/grpc/object/v2beta"
	"github.com/zitadel/zitadel/pkg/grpc/settings/v2beta"
)

func (s *Server) GetLoginSettings(ctx context.Context, req *settings.GetLoginSettingsRequest) (*settings.GetLoginSettingsResponse, error) {
	current, err := s.query.LoginPolicyByID(ctx, true, object.ResourceOwnerFromReq(ctx, req.GetCtx()), false)
	if err != nil {
		return nil, err
	}
	return &settings.GetLoginSettingsResponse{
		Settings: loginSettingsToPb(current),
		Details: &object_pb.Details{
			Sequence:      current.Sequence,
			ChangeDate:    timestamppb.New(current.ChangeDate),
			ResourceOwner: current.OrgID,
		},
	}, nil
}

func (s *Server) GetPasswordComplexitySettings(ctx context.Context, req *settings.GetPasswordComplexitySettingsRequest) (*settings.GetPasswordComplexitySettingsResponse, error) {
	current, err := s.query.PasswordComplexityPolicyByOrg(ctx, true, object.ResourceOwnerFromReq(ctx, req.GetCtx()), false)
	if err != nil {
		return nil, err
	}
	return &settings.GetPasswordComplexitySettingsResponse{
		Settings: passwordSettingsToPb(current),
		Details: &object_pb.Details{
			Sequence:      current.Sequence,
			ChangeDate:    timestamppb.New(current.ChangeDate),
			ResourceOwner: current.ResourceOwner,
		},
	}, nil
}

func (s *Server) GetBrandingSettings(ctx context.Context, req *settings.GetBrandingSettingsRequest) (*settings.GetBrandingSettingsResponse, error) {
	current, err := s.query.ActiveLabelPolicyByOrg(ctx, object.ResourceOwnerFromReq(ctx, req.GetCtx()), false)
	if err != nil {
		return nil, err
	}
	return &settings.GetBrandingSettingsResponse{
		Settings: brandingSettingsToPb(current, s.assetsAPIDomain(ctx)),
		Details: &object_pb.Details{
			Sequence:      current.Sequence,
			ChangeDate:    timestamppb.New(current.ChangeDate),
			ResourceOwner: current.ResourceOwner,
		},
	}, nil
}

func (s *Server) GetDomainSettings(ctx context.Context, req *settings.GetDomainSettingsRequest) (*settings.GetDomainSettingsResponse, error) {
	current, err := s.query.DomainPolicyByOrg(ctx, true, object.ResourceOwnerFromReq(ctx, req.GetCtx()), false)
	if err != nil {
		return nil, err
	}
	return &settings.GetDomainSettingsResponse{
		Settings: domainSettingsToPb(current),
		Details: &object_pb.Details{
			Sequence:      current.Sequence,
			ChangeDate:    timestamppb.New(current.ChangeDate),
			ResourceOwner: current.ResourceOwner,
		},
	}, nil
}

func (s *Server) GetLegalAndSupportSettings(ctx context.Context, req *settings.GetLegalAndSupportSettingsRequest) (*settings.GetLegalAndSupportSettingsResponse, error) {
	current, err := s.query.PrivacyPolicyByOrg(ctx, true, object.ResourceOwnerFromReq(ctx, req.GetCtx()), false)
	if err != nil {
		return nil, err
	}
	return &settings.GetLegalAndSupportSettingsResponse{
		Settings: legalAndSupportSettingsToPb(current),
		Details: &object_pb.Details{
			Sequence:      current.Sequence,
			ChangeDate:    timestamppb.New(current.ChangeDate),
			ResourceOwner: current.ResourceOwner,
		},
	}, nil
}

func (s *Server) GetLockoutSettings(ctx context.Context, req *settings.GetLockoutSettingsRequest) (*settings.GetLockoutSettingsResponse, error) {
	current, err := s.query.LockoutPolicyByOrg(ctx, true, object.ResourceOwnerFromReq(ctx, req.GetCtx()), false)
	if err != nil {
		return nil, err
	}
	return &settings.GetLockoutSettingsResponse{
		Settings: lockoutSettingsToPb(current),
		Details: &object_pb.Details{
			Sequence:      current.Sequence,
			ChangeDate:    timestamppb.New(current.ChangeDate),
			ResourceOwner: current.ResourceOwner,
		},
	}, nil
}

func (s *Server) GetActiveIdentityProviders(ctx context.Context, req *settings.GetActiveIdentityProvidersRequest) (*settings.GetActiveIdentityProvidersResponse, error) {
	links, err := s.query.IDPLoginPolicyLinks(ctx, object.ResourceOwnerFromReq(ctx, req.GetCtx()), &query.IDPLoginPolicyLinksSearchQuery{}, false)
	if err != nil {
		return nil, err
	}

	return &settings.GetActiveIdentityProvidersResponse{
		Details:           object.ToListDetails(links.SearchResponse),
		IdentityProviders: identityProvidersToPb(links.Links),
	}, nil
}

func (s *Server) GetGeneralSettings(ctx context.Context, _ *settings.GetGeneralSettingsRequest) (*settings.GetGeneralSettingsResponse, error) {
	langs, err := s.query.Languages(ctx)
	if err != nil {
		return nil, err
	}
	instance := authz.GetInstance(ctx)
	return &settings.GetGeneralSettingsResponse{
		SupportedLanguages: text.LanguageTagsToStrings(langs),
		DefaultOrgId:       instance.DefaultOrganisationID(),
		DefaultLanguage:    instance.DefaultLanguage().String(),
	}, nil
}
