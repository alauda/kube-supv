package cert

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"

	"github.com/pkg/errors"
)

func ValidateX509KeyPair(cert *x509.Certificate, privateKey *rsa.PrivateKey) (bool, error) {
	if privateKey == nil {
		return false, ErrNeedRSAPrivateKey
	}

	if cert == nil {
		return false, ErrNeedCertificate
	}

	publicKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return false, fmt.Errorf(`certificate's public key type is not rsa.PublicKey`)
	}
	return publicKey.N.Cmp(privateKey.N) == 0, nil
}

func ValidateCertWithCA(cert *x509.Certificate, ca *x509.Certificate) error {
	if cert == nil {
		return ErrNeedCertificate
	}

	if ca == nil {
		return ErrNeedCACertificate
	}

	pool := x509.NewCertPool()
	pool.AddCert(ca)

	if _, err := cert.Verify(x509.VerifyOptions{
		Roots:       pool,
		KeyUsages:   []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		CurrentTime: cert.NotBefore,
	}); err != nil {
		return errors.Wrap(err, `verify certificate with CA`)
	}
	return nil
}
