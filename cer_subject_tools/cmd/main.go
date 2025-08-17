package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/cer_ca_tools"
	"github.com/FISCO-BCOS/go-sdk/cer_subject_tools"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	//currentDir, _ := os.Getwd()
	//
	//certsDir := filepath.Join(currentDir, "certs")
	//
	//generator := cer_ca_tools.NewCertGenerator(certsDir)
	//
	//setupSubjectHTTP(generator)

	startTLSSubject := time.Now()
	TLSConnection()
	endTLSSubject := time.Since(startTLSSubject)
	fmt.Println("Subject VRF Time:", endTLSSubject)
}

func setupSubjectHTTP(cg *cer_ca_tools.CertGenerator) {
	log.Println("Testing CA certificate management using HTTP")

	subject := cer_subject_tools.NewSubject("http://localhost:8080")

	// 读取证书主体密钥
	subKeyPath := filepath.Join(cg.CertsDir, "subject.key")
	subKeyPEM, err := os.ReadFile(subKeyPath)
	if err != nil {
		log.Fatalf("<UNK> '%s' <UNK>: %v", subKeyPath, err)
	}

	subKeyBlock, _ := pem.Decode(subKeyPEM)

	subPrivKey, err := x509.ParsePKCS8PrivateKey(subKeyBlock.Bytes)
	if err != nil {
		log.Fatalf("<UNK> '%s' <UNK>: %v", subPrivKey, err)
	}

	// 类型断言为 *ecdsa.PrivateKey
	ecdsaPrivKey, ok := subPrivKey.(*ecdsa.PrivateKey)
	if !ok {
		log.Fatalf("CA私钥类型错误，期望*ecdsa.PrivateKey，得到%T", subPrivKey)
	}

	subject.PrivateKey = ecdsaPrivKey

	subject.PublicKey = &ecdsaPrivKey.PublicKey

	subjectInfo := pkix.Name{
		Country:            []string{"CN"},
		Province:           []string{"Beijing"},
		Locality:           []string{"Beijing"},
		Organization:       []string{"Test Client"},
		OrganizationalUnit: []string{"IT"},
		CommonName:         "Test Subject",
	}

	crtOps := cer_subject_tools.NewCRTOperations(subject)

	startRequestCert := time.Now()
	issueResponse, err := subject.RequestCertificate("ca_test_one", subjectInfo, crtOps)
	endRequestCert := time.Since(startRequestCert)
	fmt.Println("Subject Request Cert Time:", endRequestCert)
	if err != nil {
		log.Printf("匿名证书签发失败: %s", err)
	} else {
		log.Printf("匿名证书签发成功: %v", issueResponse)
	}
}

func TLSConnection() {
	log.Println("Testing VRF using TLS Connection")

	currentDir, _ := os.Getwd()
	certFile := currentDir + "/certs/tls_client.crt"
	keyFile := currentDir + "/certs/tls_client.key"
	caFile := currentDir + "/certs/tls_ca.crt"
	serverAddr := "localhost:8443"

	tlsClient := cer_subject_tools.NewTLSClient(certFile, keyFile, caFile, serverAddr)

	err := tlsClient.LoadCertificates()
	if err != nil {
		log.Fatal("error loading certificates")
	}

	err = tlsClient.Connect()
	if err != nil {
		log.Fatal("error connecting to TLS server")
	}
	defer tlsClient.Close()

	err = tlsClient.VerifyServerCertificate()
	if err != nil {
		log.Printf("server certificate warring: %v", err)
	}

	/*log.Println("Starting interactive session...")
	err = tlsClient.StartInteractiveSession()
	if err != nil {
		log.Printf("interactive session error: %v", err)
	}*/

	sessionID := "test-session-id"
	err = tlsClient.PerformVRFAuthentication(sessionID)
}
