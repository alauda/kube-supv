package cert

import (
	"errors"
	"io/fs"
)

const (
	CertificatePEMType   = "CERTIFICATE"
	RSAPrivateKeyPEMType = "RSA PRIVATE KEY"
	PublicKeyPEMType     = "PUBLIC KEY"
)

const (
	CertFileMode = fs.FileMode(0600)
	KeyFileMode  = fs.FileMode(0600)
)

var (
	ErrNeedCommonName    = errors.New("need CommonName")
	ErrNeedCertificate   = errors.New("need Certificate")
	ErrNeedCACertificate = errors.New("need CA Certificate")
	ErrNeedRSAPrivateKey = errors.New("need RSA PrivateKey")
)
