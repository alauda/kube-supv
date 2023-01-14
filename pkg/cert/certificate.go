package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math"
	"math/big"
	"net"
	"time"

	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

type CertRequest struct {
	CommonName   string
	Organization []string
	AlterNames   []string
	ExtKeyUsage  []x509.ExtKeyUsage
	Days         int
	PrivateKey   *rsa.PrivateKey
}

func SignRSACert(req CertRequest, caCert *x509.Certificate, caKey *rsa.PrivateKey) (*x509.Certificate, *rsa.PrivateKey, error) {
	if caKey == nil {
		return nil, nil, ErrNeedRSAPrivateKey
	}

	if caCert == nil {
		return nil, nil, ErrNeedCACertificate
	}

	if req.PrivateKey == nil {
		var err error
		req.PrivateKey, err = GenerateRSAKey(DefaultRSAKeyBits)
		if err != nil {
			return nil, nil, err
		}
	}

	if req.CommonName == "" {
		return nil, nil, ErrNeedCommonName
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64))
	if err != nil {
		return nil, nil, err
	}

	var dnsNames []string
	var ipAddrs []net.IP

	ipKeys := make(map[string]struct{})
	for _, name := range req.AlterNames {
		if ip := net.ParseIP(name); ip != nil {
			ipStr := ip.String()
			if _, ok := ipKeys[ipStr]; !ok {
				ipKeys[ipStr] = struct{}{}
				ipAddrs = append(ipAddrs, ip)
			}
		} else {
			if !funk.ContainsString(dnsNames, name) {
				dnsNames = append(dnsNames, name)
			}
		}
	}

	now := time.Now()

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   req.CommonName,
			Organization: req.Organization,
		},
		DNSNames:              dnsNames,
		IPAddresses:           ipAddrs,
		SerialNumber:          serial,
		NotBefore:             now.Add(time.Duration(-24) * time.Hour).UTC(),
		NotAfter:              now.Add(time.Duration(req.Days) * time.Hour).UTC(),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           req.ExtKeyUsage,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, &certTmpl, caCert, req.PrivateKey.Public(), caKey)
	if err != nil {
		return nil, nil, errors.Wrapf(err, `create certificate for signed certificate "%s"`, req.CommonName)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, errors.Wrapf(err, `parse DER data for signed certificate for "%s"`, req.CommonName)
	}
	return cert, req.PrivateKey, nil
}
