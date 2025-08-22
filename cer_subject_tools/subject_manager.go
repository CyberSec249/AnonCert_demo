package cer_subject_tools

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"time"
)

const defaultServerURL = "http://localhost:8080"

type Subject struct {
	SubjectURL string
	PublicKey  *ecdsa.PublicKey  `json:"public_key"`
	PrivateKey *ecdsa.PrivateKey `json:"private_key"`
}

type CertificateRequest struct {
	SubjectInfo        pkix.Name `json:"subject"`
	PublicKeyAlgorithm int       `json:"public_key_algorithm"`
	PublicKeyBytes     []byte    `json:"public_key_bytes"`
	SignatureAlgorithm int       `json:"signature_algorithm"`
}

type CertificateResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Certificate string `json:"certificate"`
	Err         error  `json:"error,omitempty"`
}

type HTTPCertRevokeRequest struct {
	SerialNumber string `json:"serial_number"`
	Reason       int    `json:"reason"`
}

func NewSubject(SubjectURL string) *Subject {
	if SubjectURL == "" {
		SubjectURL = "http://localhost:8080"
	}
	return &Subject{
		SubjectURL: SubjectURL,
	}
}

func (s *Subject) CreateCertIssueRequest(subjectInfo pkix.Name) (*x509.CertificateRequest, error) {
	if s.PrivateKey == nil {
		return nil, fmt.Errorf("private key not provided")
	}

	template := x509.CertificateRequest{
		Version:            3,
		Subject:            subjectInfo,
		PublicKeyAlgorithm: 3,
		PublicKey:          s.PublicKey,
		Signature:          nil,
		SignatureAlgorithm: 10,
		Extensions:         nil,
	}

	cirDER, err := x509.CreateCertificateRequest(rand.Reader, &template, s.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to create CSR: %s", err)
	}

	cir, err := x509.ParseCertificateRequest(cirDER)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse CSR: %s", err)
	}

	return cir, nil
}

func (s *Subject) SendCertificateIssueRequest(caName string, cir *x509.CertificateRequest, xorResult []byte, remainders []*big.Int) (*CertificateResponse, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(cir.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize public key: %s", err)
	}

	anonCertRequest := struct {
		SubjectInfo        pkix.Name  `json:"subject"`
		PublicKeyAlgorithm int        `json:"public_key_algorithm"`
		PublicKeyBytes     []byte     `json:"public_key_bytes"`
		SignatureAlgorithm int        `json:"signature_algorithm"`
		XORResult          []byte     `json:"xor_result"`
		Remainders         []*big.Int `json:"remainders"`
	}{
		SubjectInfo:        cir.Subject,
		PublicKeyAlgorithm: int(cir.PublicKeyAlgorithm),
		PublicKeyBytes:     publicKeyBytes,
		SignatureAlgorithm: int(cir.SignatureAlgorithm),
		XORResult:          xorResult,
		Remainders:         remainders,
	}

	jsonData, err := json.Marshal(anonCertRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal certificate issue request: %w", err)
	}

	url := fmt.Sprintf("%s/certificate/issue?caName=%s", defaultServerURL, caName)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send certificate request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send certificate request: %s", string(body))
	}

	var certificateResponse CertificateResponse
	if err = json.Unmarshal(body, &certificateResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return &certificateResponse, nil
}

func (s *Subject) RequestCertificate(caName string, subjectInfo pkix.Name, crtOps *CRTOperations) (*CertificateResponse, error) {
	cir, err := s.CreateCertIssueRequest(subjectInfo)

	caURLs := []string{"http://localhost:8080", "http://localhost:8080", "http://localhost:8080"}

	caNames := []string{"ca_test_one", "ca_test_two", "ca_test_three"}
	if cir != nil {
		startCRTGeneration := time.Now()
		xorResult := CRTGeneration(caURLs, caNames, cir.Subject, crtOps)
		endCRTGeneration := time.Since(startCRTGeneration)
		fmt.Println("Subject CRT Generation Time:", endCRTGeneration)
		return s.SendCertificateIssueRequest(caName, cir, xorResult, crtOps.Remainders)
	} else {
		return nil, err
	}
}

func (s *Subject) RequestRevokeCertificate(caName string, serialNumber string, reason int) (*CertificateResponse, error) {
	revokeRequest := HTTPCertRevokeRequest{
		SerialNumber: serialNumber,
		Reason:       reason,
	}

	jsonData, err := json.Marshal(revokeRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal revoke request: %w", err)
	}

	url := fmt.Sprintf("%s/certificate/revoke?caName=%s", defaultServerURL, caName)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send revoke request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to send revoke request: %s", string(body))
	}

	var revokeResponse CertificateResponse
	if err = json.Unmarshal(body, &revokeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}
	return &revokeResponse, nil
}

func CRTGeneration(caURLs []string, caNames []string, subjectInfo pkix.Name, crtOps *CRTOperations) []byte {
	err := crtOps.RequestAllModuli(caURLs, caNames, subjectInfo.CommonName)
	if err != nil {
		log.Fatalf("请求模数时出错: %v", err)
	}

	for i, modulus := range crtOps.Moduli {
		fmt.Printf("模数 %d: %s\n", i+1, modulus.String())
	}

	err = crtOps.GenerateRandomRemainders()
	if err != nil {
		log.Fatalf("生成随机余数时出错: %v", err)
	}

	for i, remainder := range crtOps.Remainders {
		fmt.Printf("余数 %d: %s\n", i+1, remainder.String())
	}

	err = crtOps.SolveChineseRemainderTheorem()
	if err != nil {
		log.Fatalf("计算中国剩余定理时出错: %v", err)
	}

	fmt.Printf("计算结果x: %s\n", crtOps.X.String())

	subjectInfoBytes, err := json.Marshal(subjectInfo)
	if err != nil {
		log.Fatalf("序列化主体信息时出错: %v", err)
	}

	xorResult, err := crtOps.XORWithSubjectInfo(subjectInfoBytes)
	if err != nil {
		log.Fatalf("进行异或操作时出错: %v", err)
	}

	return xorResult
}
