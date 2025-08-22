package cer_subject_tools

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"math/big"
	"time"
)

func PrintCertInfo(cert *x509.Certificate) {
	fmt.Println("-----BEGIN CERTIFICATE-----")
	fmt.Println("Version: ", cert.Version)
	fmt.Println("SerialNumber: ", cert.SerialNumber)
	fmt.Println("Issuer:", cert.Issuer)
	fmt.Println("Subject: ", cert.Subject)
	fmt.Println("PublicKeyAlgorithm: ", cert.PublicKeyAlgorithm)
	fmt.Println("Subject's PublicKey", cert.PublicKey)
	fmt.Println("NotBefore: ", cert.NotBefore)
	fmt.Println("NotAfter: ", cert.NotAfter)
	fmt.Println("SignatureAlgorithm: ", cert.SignatureAlgorithm)
	fmt.Println("Signature: ", cert.Signature)
	fmt.Println("-----END CERTIFICATE-----")
}

func GenerateCert(isCA bool, caPrivateKey *ecdsa.PrivateKey, caCert *x509.Certificate, subjectPublicKey *ecdsa.PublicKey, subject pkix.Name, issuer pkix.Name, days int) (*x509.Certificate, []byte, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	SerialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, nil, err
	}

	notBefore := time.Now()
	notAfter := notBefore.Add(time.Duration(days) * 24 * time.Hour)
	certTemplate := x509.Certificate{
		Version:            3,
		SerialNumber:       SerialNumber,
		Issuer:             issuer,
		Subject:            subject,
		PublicKeyAlgorithm: 3,
		PublicKey:          subjectPublicKey,
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		Extensions:         nil,
		SignatureAlgorithm: 10,
		Signature:          nil,
		IsCA:               isCA,
	}

	if isCA {
		certTemplate.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		certTemplate.BasicConstraintsValid = true
		certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &caPrivateKey.PublicKey, caPrivateKey)
		cert, err := x509.ParseCertificate(certDER)
		return cert, certDER, err
	} else {
		certTemplate.KeyUsage = x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature
		certDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCert, subjectPublicKey, caPrivateKey)
		cert, err := x509.ParseCertificate(certDER)
		return cert, certDER, err
	}
}
