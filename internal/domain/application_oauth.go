package domain

import (
	"github.com/zitadel/zitadel/internal/id"
)

type oAuthApplication interface {
	setClientID(clientID string)
	setClientSecret(encodedHash string)
	requiresClientSecret() bool
}

// ClientID random_number (eg. 495894098234)
func SetNewClientID(a oAuthApplication, idGenerator id.Generator) error {
	clientID, err := idGenerator.Next()
	if err != nil {
		return err
	}

	a.setClientID(clientID)
	return nil
}

func SetNewClientSecretIfNeeded(a oAuthApplication, generate func() (encodedHash, plain string, err error)) (string, error) {
	if !a.requiresClientSecret() {
		return "", nil
	}
	encodedHash, plain, err := generate()
	if err != nil {
		return "", err
	}
	a.setClientSecret(encodedHash)
	return plain, nil
}
