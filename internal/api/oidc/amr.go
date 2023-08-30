package oidc

import "github.com/zitadel/zitadel/internal/domain"

const (
	// Password states that the users password has been verified
	// Deprecated: use `PWD` instead
	Password = "password"
	// PWD states that the users password has been verified
	PWD = "pwd"
	// MFA states that multiple factors have been verified (e.g. pwd and otp or passkey)
	MFA = "mfa"
	// OTP states that a one time password has been verified (e.g. TOTP)
	OTP = "otp"
	// UserPresence states that the end users presence has been verified (e.g. passkey and u2f)
	UserPresence = "user"
)

// AuthMethodTypesToAMR maps zitadel auth method types to Authentication Method Reference Values
// as defined in [RFC 8176, section 2].
//
// [RFC 8176, section 2]: https://datatracker.ietf.org/doc/html/rfc8176#section-2
func AuthMethodTypesToAMR(methodTypes []domain.UserAuthMethodType) []string {
	amr := make([]string, 0, 4)
	var factors, otp int
	for _, methodType := range methodTypes {
		switch methodType {
		case domain.UserAuthMethodTypePassword:
			amr = append(amr, PWD)
			factors++
		case domain.UserAuthMethodTypePasswordless:
			amr = append(amr, UserPresence)
			factors += 2
		case domain.UserAuthMethodTypeU2F:
			amr = append(amr, UserPresence)
			factors++
		case domain.UserAuthMethodTypeTOTP,
			domain.UserAuthMethodTypeOTPSMS,
			domain.UserAuthMethodTypeOTPEmail:
			// a user could use multiple (t)otp, which is a factor, but still will be returned as a single `otp` entry
			otp++
			factors++
		case domain.UserAuthMethodTypeIDP:
			// no AMR value according to specification
			factors++
		case domain.UserAuthMethodTypeUnspecified:
			// ignore
		}
	}
	if otp > 0 {
		amr = append(amr, OTP)
	}
	if factors >= 2 {
		amr = append(amr, MFA)
	}
	return amr
}
