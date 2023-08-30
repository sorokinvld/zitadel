package crypto

import (
	"crypto/rand"
	"time"

	"github.com/zitadel/zitadel/internal/errors"
)

var (
	lowerLetters = []rune("abcdefghijklmnopqrstuvwxyz")
	upperLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	digits       = []rune("0123456789")
	symbols      = []rune("~!@#$^&*()_+`-={}|[]:<>?,./")
)

type GeneratorConfig struct {
	Length              uint
	Expiry              time.Duration
	IncludeLowerLetters bool
	IncludeUpperLetters bool
	IncludeDigits       bool
	IncludeSymbols      bool
}

type Generator interface {
	Length() uint
	Expiry() time.Duration
	Alg() Crypto
	Runes() []rune
}

type generator struct {
	length uint
	expiry time.Duration
	runes  []rune
}

func (g *generator) Length() uint {
	return g.length
}

func (g *generator) Expiry() time.Duration {
	return g.expiry
}

func (g *generator) Runes() []rune {
	return g.runes
}

type encryptionGenerator struct {
	generator
	alg EncryptionAlgorithm
}

func (g *encryptionGenerator) Alg() Crypto {
	return g.alg
}

func NewEncryptionGenerator(config GeneratorConfig, algorithm EncryptionAlgorithm) Generator {
	return &encryptionGenerator{
		newGenerator(config),
		algorithm,
	}
}

type hashGenerator struct {
	generator
	alg HashAlgorithm
}

func (g *hashGenerator) Alg() Crypto {
	return g.alg
}

func NewHashGenerator(config GeneratorConfig, algorithm HashAlgorithm) Generator {
	return &hashGenerator{
		newGenerator(config),
		algorithm,
	}
}

func newGenerator(config GeneratorConfig) generator {
	var runes []rune
	if config.IncludeLowerLetters {
		runes = append(runes, lowerLetters...)
	}
	if config.IncludeUpperLetters {
		runes = append(runes, upperLetters...)
	}
	if config.IncludeDigits {
		runes = append(runes, digits...)
	}
	if config.IncludeSymbols {
		runes = append(runes, symbols...)
	}
	return generator{
		length: config.Length,
		expiry: config.Expiry,
		runes:  runes,
	}
}

func NewCode(g Generator) (*CryptoValue, string, error) {
	code, err := GenerateRandomString(g.Length(), g.Runes())
	if err != nil {
		return nil, "", err
	}
	crypto, err := Crypt([]byte(code), g.Alg())
	if err != nil {
		return nil, "", err
	}
	return crypto, code, nil
}

func IsCodeExpired(creationDate time.Time, expiry time.Duration) bool {
	if expiry == 0 {
		return false
	}
	return creationDate.Add(expiry).Before(time.Now().UTC())
}

func VerifyCode(creationDate time.Time, expiry time.Duration, cryptoCode *CryptoValue, verificationCode string, g Generator) error {
	return VerifyCodeWithAlgorithm(creationDate, expiry, cryptoCode, verificationCode, g.Alg())
}

func VerifyCodeWithAlgorithm(creationDate time.Time, expiry time.Duration, cryptoCode *CryptoValue, verificationCode string, algorithm Crypto) error {
	if IsCodeExpired(creationDate, expiry) {
		return errors.ThrowPreconditionFailed(nil, "CODE-QvUQ4P", "Errors.User.Code.Expired")
	}
	switch alg := algorithm.(type) {
	case EncryptionAlgorithm:
		return verifyEncryptedCode(cryptoCode, verificationCode, alg)
	case HashAlgorithm:
		return verifyHashedCode(cryptoCode, verificationCode, alg)
	}
	return errors.ThrowInvalidArgument(nil, "CODE-fW2gNa", "Errors.User.Code.GeneratorAlgNotSupported")
}

func GenerateRandomString(length uint, chars []rune) (string, error) {
	if length == 0 {
		return "", nil
	}

	max := len(chars) - 1
	maxStr := int(length - 1)

	str := make([]rune, length)
	randBytes := make([]byte, length)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	for i, rb := range randBytes {
		str[i] = chars[int(rb)%max]
		if i == maxStr {
			return string(str), nil
		}
	}
	return "", nil
}

func verifyEncryptedCode(cryptoCode *CryptoValue, verificationCode string, alg EncryptionAlgorithm) error {
	if cryptoCode == nil {
		return errors.ThrowInvalidArgument(nil, "CRYPT-aqrFV", "Errors.User.Code.CryptoCodeNil")
	}
	code, err := DecryptString(cryptoCode, alg)
	if err != nil {
		return err
	}

	if code != verificationCode {
		return errors.ThrowInvalidArgument(nil, "CODE-woT0xc", "Errors.User.Code.Invalid")
	}
	return nil
}

func verifyHashedCode(cryptoCode *CryptoValue, verificationCode string, alg HashAlgorithm) error {
	if cryptoCode == nil {
		return errors.ThrowInvalidArgument(nil, "CRYPT-2q3r", "cryptoCode must not be nil")
	}
	return CompareHash(cryptoCode, []byte(verificationCode), alg)
}
