package cert

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/alauda/kube-supv/pkg/utils"
	"github.com/pkg/errors"
)

func SaveCert(cert *x509.Certificate, path string) error {
	if cert == nil {
		return ErrNeedCertificate
	}

	file, err := utils.OpenFileToWrite(path, CertFileMode)
	if err != nil {
		return err
	}

	if err := pem.Encode(file, &pem.Block{
		Type:  CertificatePEMType,
		Bytes: cert.Raw,
	}); err != nil {
		file.Close()
		return errors.Wrapf(err, `pem encode "%s"`, path)
	}

	if err := file.Close(); err != nil {
		return errors.Wrapf(err, `close "%s"`, path)
	}

	return nil
}

func CertEncodeToBytes(cert *x509.Certificate) ([]byte, error) {
	if cert == nil {
		return nil, ErrNeedCertificate
	}

	buf := bytes.Buffer{}

	if err := pem.Encode(&buf, &pem.Block{
		Type:  CertificatePEMType,
		Bytes: cert.Raw,
	}); err != nil {
		return nil, errors.Wrap(err, `pem encode certificate`)
	}

	return buf.Bytes(), nil
}

func SaveRSAPrivateKey(privateKey *rsa.PrivateKey, path string) error {
	if privateKey == nil {
		return ErrNeedRSAPrivateKey
	}

	file, err := utils.OpenFileToWrite(path, KeyFileMode)
	if err != nil {
		return err
	}

	if err := pem.Encode(file, &pem.Block{
		Type:  RSAPrivateKeyPEMType,
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		file.Close()
		return errors.Wrapf(err, `pem encode rsa private key to "%s"`, path)
	}

	if err := file.Close(); err != nil {
		return errors.Wrapf(err, `close "%s"`, path)
	}

	return nil
}

func RSAPrivateKeyEncodeToBytes(privateKey *rsa.PrivateKey) ([]byte, error) {
	if privateKey == nil {
		return nil, ErrNeedRSAPrivateKey
	}

	buf := bytes.Buffer{}

	if err := pem.Encode(&buf, &pem.Block{
		Type:  RSAPrivateKeyPEMType,
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}); err != nil {
		return nil, errors.Wrap(err, `pem encode rsa private key`)
	}

	return buf.Bytes(), nil
}

func SavePublickKey(privateKey *rsa.PrivateKey, path string) error {
	if privateKey == nil {
		return ErrNeedRSAPrivateKey
	}

	der, err := x509.MarshalPKIXPublicKey(privateKey.Public())
	if err != nil {
		return errors.Wrapf(err, `marshal public key to "%s"`, path)
	}

	file, err := utils.OpenFileToWrite(path, CertFileMode)
	if err != nil {
		return err
	}

	if err := pem.Encode(file, &pem.Block{
		Type:  PublicKeyPEMType,
		Bytes: der,
	}); err != nil {
		file.Close()
		return errors.Wrapf(err, `pem encode public key to "%s"`, path)
	}

	if err := file.Close(); err != nil {
		return errors.Wrapf(err, `close "%s"`, path)
	}

	return nil
}

func LoadCert(path string) (*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, `load cert from "%s"`, path)
	}

	var block *pem.Block
	for {
		block, data = pem.Decode(data)
		if block == nil {
			return nil, fmt.Errorf(`con not find validate cert from "%s"`, path)
		}
		if block.Type == CertificatePEMType && len(block.Headers) == 0 {
			break
		}
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, `parse cert from "%s"`, path)
	}
	return cert, nil
}

func LoadRSAPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, `load ras private Key from "%s"`, path)
	}

	var block *pem.Block
	for {
		block, data = pem.Decode(data)
		if block == nil {
			return nil, fmt.Errorf(`con not find rsa private key from "%s"`, path)
		}

		if block.Type == RSAPrivateKeyPEMType {
			break
		}
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, `parse rsa private key from "%s"`, path)
	}
	return key, nil
}

func LoadRSAPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, `load public Key from "%s"`, path)
	}

	var block *pem.Block
	for {
		block, data = pem.Decode(data)
		if block == nil {
			return nil, fmt.Errorf(`con not find public key from "%s"`, path)
		}

		if block.Type == PublicKeyPEMType {
			break
		}
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.Wrapf(err, `parse public key from "%s"`, path)
	}
	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.Wrapf(err, `"%s" is not rsa public key`, path)
	}
	return publicKey, nil
}

func CanLoadFiles(files ...string) (bool, error) {
	if len(files) == 0 {
		return false, nil
	}

	exists := make(map[string]bool, len(files))
	for _, file := range files {
		var err error
		exists[file], err = utils.IsFileExist(file)
		if err != nil {
			return false, err
		}
	}

	allExist := true
	notExistFile := ""

	allNotExist := true
	existFile := ""

	for file, exist := range exists {
		if exist {
			allNotExist = false
			existFile = file
		} else {
			allExist = false
			notExistFile = file
		}
	}

	if (!allExist) && (!allNotExist) {
		return false, fmt.Errorf(`"%s" exists but "%s" doest not exist`, existFile, notExistFile)
	}

	return allExist, nil
}

func LoadCertKeyPair(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	cert, err := LoadCert(certPath)
	if err != nil {
		return nil, nil, err
	}
	key, err := LoadRSAPrivateKey(keyPath)
	if err != nil {
		return nil, nil, err
	}

	match, err := ValidateX509KeyPair(cert, key)
	if err != nil {
		return nil, nil, errors.Wrapf(err, `validate certificate and key for "%s" and "%s"`, certPath, keyPath)
	}
	if !match {
		return nil, nil, fmt.Errorf(`"%s" and "%s" do not match`, certPath, keyPath)
	}
	return cert, key, nil
}

func LoadRSAKeyPair(publicKeyPath, privateKeyPath string) (*rsa.PublicKey, *rsa.PrivateKey, error) {
	publicKey, err := LoadRSAPublicKey(publicKeyPath)
	if err != nil {
		return nil, nil, err
	}
	privateKey, err := LoadRSAPrivateKey(privateKeyPath)
	if err != nil {
		return nil, nil, err
	}

	if privateKey.E != publicKey.E || privateKey.N.Cmp(publicKey.N) != 0 {
		return nil, nil, fmt.Errorf(`"%s" and "%s" do not match`, publicKeyPath, privateKeyPath)

	}

	return publicKey, privateKey, nil
}
