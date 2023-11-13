package user

import (
	"context"

	"github.com/zitadel/zitadel/internal/api/authz"
	"github.com/zitadel/zitadel/internal/api/grpc/object/v2"
	"github.com/zitadel/zitadel/internal/domain"
	caos_errs "github.com/zitadel/zitadel/internal/errors"
	user "github.com/zitadel/zitadel/pkg/grpc/user/v2beta"
)

func (s *Server) PasswordReset(ctx context.Context, req *user.PasswordResetRequest) (_ *user.PasswordResetResponse, err error) {
	var details *domain.ObjectDetails
	var code *string

	switch m := req.GetMedium().(type) {
	case *user.PasswordResetRequest_SendLink:
		details, code, err = s.command.RequestPasswordResetURLTemplate(ctx, req.GetUserId(), m.SendLink.GetUrlTemplate(), notificationTypeToDomain(m.SendLink.GetNotificationType()))
	case *user.PasswordResetRequest_ReturnCode:
		details, code, err = s.command.RequestPasswordResetReturnCode(ctx, req.GetUserId())
	case nil:
		details, code, err = s.command.RequestPasswordReset(ctx, req.GetUserId())
	default:
		err = caos_errs.ThrowUnimplementedf(nil, "USERv2-SDeeg", "verification oneOf %T in method RequestPasswordReset not implemented", m)
	}
	if err != nil {
		return nil, err
	}

	return &user.PasswordResetResponse{
		Details:          object.DomainToDetailsPb(details),
		VerificationCode: code,
	}, nil
}

func notificationTypeToDomain(notificationType user.NotificationType) domain.NotificationType {
	switch notificationType {
	case user.NotificationType_NOTIFICATION_TYPE_Email:
		return domain.NotificationTypeEmail
	case user.NotificationType_NOTIFICATION_TYPE_SMS:
		return domain.NotificationTypeSms
	case user.NotificationType_NOTIFICATION_TYPE_Unspecified:
		return domain.NotificationTypeEmail
	default:
		return domain.NotificationTypeEmail
	}
}

func (s *Server) SetPassword(ctx context.Context, req *user.SetPasswordRequest) (_ *user.SetPasswordResponse, err error) {
	var resourceOwner = authz.GetCtxData(ctx).OrgID
	var details *domain.ObjectDetails

	switch v := req.GetVerification().(type) {
	case *user.SetPasswordRequest_CurrentPassword:
		details, err = s.command.ChangePassword(ctx, resourceOwner, req.GetUserId(), v.CurrentPassword, req.GetNewPassword().GetPassword(), "")
	case *user.SetPasswordRequest_VerificationCode:
		details, err = s.command.SetPasswordWithVerifyCode(ctx, resourceOwner, req.GetUserId(), v.VerificationCode, req.GetNewPassword().GetPassword(), "")
	case nil:
		details, err = s.command.SetPassword(ctx, resourceOwner, req.GetUserId(), req.GetNewPassword().GetPassword(), req.GetNewPassword().GetChangeRequired())
	default:
		err = caos_errs.ThrowUnimplementedf(nil, "USERv2-SFdf2", "verification oneOf %T in method SetPasswordRequest not implemented", v)
	}
	if err != nil {
		return nil, err
	}

	return &user.SetPasswordResponse{
		Details: object.DomainToDetailsPb(details),
	}, nil
}
