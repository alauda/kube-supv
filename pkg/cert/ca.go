package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"

	"github.com/pkg/errors"
)

func GenerateSelfSignedRSACA(cn string, days int) (*x509.Certificate, *rsa.PrivateKey, error) {
	if cn == "" {
		return nil, nil, ErrNeedCommonName
	}

	privateKey, err := GenerateRSAKey(DefaultRSACAKeyBits)
	if err != nil {
		return nil, nil, errors.Wrapf(err, `generate private key for self-signed ca "%s"`, cn)
	}

	now := time.Now()
	template := x509.Certificate{
		SerialNumber: new(big.Int).SetInt64(0),
		Subject: pkix.Name{
			CommonName: cn,
		},
		NotBefore:             now.Add(time.Duration(-24) * time.Hour).UTC(),
		NotAfter:              now.Add(time.Duration(days*24) * time.Hour).UTC(),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, privateKey.Public(), privateKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, `create certificate for self-signed ca "%s"`, cn)
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, errors.Wrapf(err, `parse DER data for self-signed ca for "%s"`, cn)
	}

	return cert, privateKey, nil
}
