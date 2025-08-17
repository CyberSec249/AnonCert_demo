package cer_ca_tools

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/cer_subject_tools"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CA struct {
	Name           pkix.Name                           `json:"name"`
	PublicKey      *ecdsa.PublicKey                    `json:"public_key"`
	PrivateKey     *ecdsa.PrivateKey                   `json:"private_key"`
	Certificate    *x509.Certificate                   `json:"certificate"`
	CertificatePEM []byte                              `json:"certificate_pem"`
	IssuedCerts    map[string]*x509.Certificate        `json:"issued_certs"`
	RevokedCerts   map[string]*pkix.RevokedCertificate `json:"revoked_certs"`
	Mutex          sync.Mutex                          `json:"-"`
}

type CertificateRequest struct {
	Subject            pkix.Name `json:"subject"`
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

type CAManager struct {
	CAs       map[string]*CA
	PrimePool *PrimePool
	mutex     sync.RWMutex
}

func NewCAManager() *CAManager {
	return &CAManager{
		CAs:       make(map[string]*CA),
		PrimePool: NewPrimePool(),
		mutex:     sync.RWMutex{},
	}
}

func CreateNewCA(caName string) (*CA, error) {
	caSK, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("Error generating CA certificate: %s", err)
	}

	ca := pkix.Name{
		CommonName:         caName,
		Organization:       []string{"xidian"},
		OrganizationalUnit: []string{"Acme Co"},
		Country:            []string{"CN"},
		Province:           []string{"xi'an"},
		Locality:           nil,
	}
	//CA self-signature certificate
	caCert, caCertDER, err := cer_subject_tools.GenerateCert(true, caSK, nil, &caSK.PublicKey, ca, ca, 365)
	if err != nil {
		fmt.Println("failed to generate CA cert:", err)
	}

	currentDir, _ := os.Getwd()

	certsDir := filepath.Join(currentDir, "certs", "ca")

	caCertPath := filepath.Join(certsDir, caName+".crt")
	caCertFile, err := os.Create(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("Error creating CA certificate file: %s", err)
	}
	defer caCertFile.Close()

	err = pem.Encode(caCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	})
	if err != nil {
		return nil, fmt.Errorf("Error encoding CA certificate: %s", err)
	}

	caKeyPath := filepath.Join(certsDir, caName+".key")
	caKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Error creating CA key file: %s", err)
	}
	defer caKeyFile.Close()

	caPKDER, err := x509.MarshalPKCS8PrivateKey(caSK)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling CA private key: %s", err)
	}

	err = pem.Encode(caKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caPKDER,
	})
	if err != nil {
		return nil, fmt.Errorf("Error encoding CA private key: %s", err)
	}

	return &CA{
		Name:           ca,
		PublicKey:      &caSK.PublicKey,
		PrivateKey:     caSK,
		Certificate:    caCert,
		CertificatePEM: caCertDER,
		IssuedCerts:    make(map[string]*x509.Certificate),
		RevokedCerts:   make(map[string]*pkix.RevokedCertificate),
		Mutex:          sync.Mutex{},
	}, nil
}

func (manager *CAManager) AddCAToManager(ca *CA) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	manager.CAs[ca.Name.CommonName] = ca

	log.Println("Added CA to manager", ca.Name)
}

func (manager *CAManager) GetCAInfo(caName string) (*CA, bool) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	ca, exists := manager.CAs[caName]
	return ca, exists
}

func (manager *CAManager) SetupHTTPHandlers() {
	http.HandleFunc("/certificate/issue", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		startRequestCertToCA := time.Now()

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 解析带有XOR结果的证书请求
		var anonCertRequest struct {
			SubjectInfo        pkix.Name  `json:"subject"`
			PublicKeyAlgorithm int        `json:"public_key_algorithm"`
			PublicKeyBytes     []byte     `json:"public_key_bytes"`
			SignatureAlgorithm int        `json:"signature_algorithm"`
			XORResult          []byte     `json:"xor_result"`
			Remainders         []*big.Int `json:"remainders"`
		}

		if err := json.Unmarshal(body, &anonCertRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		caName := r.URL.Query().Get("caName")
		if caName == "" {
			http.Error(w, "caName is required", http.StatusBadRequest)
			return
		}

		ca, exists := manager.CAs[caName]
		if !exists {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		subjectPK, err := x509.ParsePKIXPublicKey(anonCertRequest.PublicKeyBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ecdsaPublicKey, ok := subjectPK.(*ecdsa.PublicKey)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !manager.VerifyXORResult(anonCertRequest.XORResult, anonCertRequest.SubjectInfo, anonCertRequest.Remainders) {
			http.Error(w, "主体信息异或逆运算验证失败", http.StatusInternalServerError)
			return
		}

		xorHash := sha256.Sum256(anonCertRequest.XORResult)
		anonymousSubject := pkix.Name{
			CommonName:         fmt.Sprintf("anonymous-%x", xorHash[:16]),
			Organization:       []string{"Anonymous Organization"},
			OrganizationalUnit: []string{"Anonymous Department"},
			Country:            []string{"AN"},
			Province:           []string{"Anonymous Province"},
			Locality:           []string{"Anonymous Locality"},
		}

		response := ca.IssueCertificate(anonymousSubject, ecdsaPublicKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		endRequestCertToCA := time.Since(startRequestCertToCA)
		fmt.Println("CA Issue Cert Time:", endRequestCertToCA)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/certificate/revoke", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var httpCertRevokeRequest HTTPCertRevokeRequest
		if err := json.Unmarshal(body, &httpCertRevokeRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		caName := r.URL.Query().Get("caName")
		if caName == "" {
			http.Error(w, "HTTP请求缺少CA名称", http.StatusBadRequest)
			return
		}
		issuerCA, _, err := manager.FindCertIssuer(httpCertRevokeRequest.SerialNumber)
		if err != nil {
			fmt.Errorf("Error finding issuer CA: %s", err)
		}

		response := issuerCA.RevokeCertificate(caName, httpCertRevokeRequest.SerialNumber, httpCertRevokeRequest.Reason)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// 添加模数请求处理
	http.HandleFunc("/certificate/modulus/request", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			return
		}

		caName := r.URL.Query().Get("caName")
		if caName == "" {
			http.Error(w, "缺少CA名称", http.StatusBadRequest)
			return
		}

		_, exists := manager.GetCAInfo(caName)
		if !exists {
			http.Error(w, "CA不存在", http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体时出错: %s", err), http.StatusBadRequest)
			return
		}

		var modulusRequest struct {
			SubjectID string `json:"subject_id"`
		}

		if err := json.Unmarshal(body, &modulusRequest); err != nil {
			http.Error(w, fmt.Sprintf("解析请求体时出错: %s", err), http.StatusBadRequest)
			return
		}

		// 从质数池中随机选择一个质数
		prime, err := manager.PrimePool.GetRandomPrime()
		if err != nil {
			http.Error(w, fmt.Sprintf("获取质数时出错: %s", err), http.StatusInternalServerError)
			return
		}

		// 返回模数
		response := struct {
			Success bool     `json:"success"`
			Message string   `json:"message"`
			Modulus *big.Int `json:"modulus"`
		}{
			Success: true,
			Message: "成功获取模数",
			Modulus: prime,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

		log.Printf("CA %s 为主体 %s 提供了模数 %s", caName, modulusRequest.SubjectID, prime.String())
	})

	// 添加XOR结果处理
	http.HandleFunc("/certificate/modulus/xor", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
			return
		}

		caName := r.URL.Query().Get("caName")
		if caName == "" {
			http.Error(w, "HTTP请求缺少CA名称", http.StatusBadRequest)
			return
		}

		_, exists := manager.GetCAInfo(caName)
		if !exists {
			http.Error(w, "CA不存在", http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体时出错: %s", err), http.StatusBadRequest)
			return
		}

		var xorRequest struct {
			SubjectID  string   `json:"subject_id"`
			XORResult  []byte   `json:"xor_result"`
			Remainders [][]byte `json:"remainders"`
		}

		if err := json.Unmarshal(body, &xorRequest); err != nil {
			http.Error(w, fmt.Sprintf("解析请求体时出错: %s", err), http.StatusBadRequest)
			return
		}

		// 在实际应用中，CA会存储或处理这些信息
		// 这里我们只记录日志
		log.Printf("收到主体 %s 的XOR结果，长度为 %d 字节，包含 %d 个余数",
			xorRequest.SubjectID, len(xorRequest.XORResult), len(xorRequest.Remainders))

		// 返回成功响应
		response := struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
		}{
			Success: true,
			Message: "成功接收XOR结果",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
}

func (ca *CA) IssueCertificate(subject pkix.Name, subjectPublicKey *ecdsa.PublicKey) CertificateResponse {
	ca.Mutex.Lock()
	defer ca.Mutex.Unlock()

	log.Printf("CA %s is issuing certificate", ca.Name.CommonName)

	//生成随机序列号
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 160)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	// 创建服务器证书模板
	serverTemplate := x509.Certificate{
		SerialNumber:          serialNumber,
		Subject:               subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1年
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	subjectCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, ca.Certificate, subjectPublicKey, ca.PrivateKey)
	if err != nil {
		log.Printf("生成服务器证书失败: %v", err)
	}

	currentDir, _ := os.Getwd()
	certsDir := filepath.Join(currentDir, "certs")
	subjectCertPath := filepath.Join(certsDir, "subject.crt")
	subjectCertFile, err := os.Create(subjectCertPath)
	if err != nil {
		log.Printf("创建服务器证书文件失败: %v", err)
	}
	defer subjectCertFile.Close()
	err = pem.Encode(subjectCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: subjectCertDER,
	})
	if err != nil {
		fmt.Errorf("Error encoding server certificate: %s", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: subjectCertDER,
	})

	subjectCert, err := x509.ParseCertificate(subjectCertDER)
	if err != nil {
		log.Fatalf("解析证书失败: %v", err)
	}

	srtialStr := serialNumber.String()
	ca.IssuedCerts[srtialStr] = subjectCert

	log.Printf("issue certificate success for CA: %s, Serial: %s", ca.Name.CommonName, srtialStr)
	return CertificateResponse{
		Success:     true,
		Message:     "certificate issued",
		Certificate: string(certPEM),
	}
}

func (ca *CA) RevokeCertificate(caName string, serialNumber string, reason int) CertificateResponse {
	ca.Mutex.Lock()
	defer ca.Mutex.Unlock()

	cert, exists := ca.IssuedCerts[serialNumber]
	if !exists {
		return CertificateResponse{
			Success: false,
			Message: "撤销失败，证书不存在",
		}
	}

	if _, revoked := ca.RevokedCerts[serialNumber]; revoked {
		return CertificateResponse{
			Success: false,
			Message: "撤销失败，证书已经被撤销",
		}
	}

	revokedCert := pkix.RevokedCertificate{
		SerialNumber:   cert.SerialNumber,
		RevocationTime: time.Now(),
		Extensions: []pkix.Extension{
			{
				Id:    []int{2, 5, 29, 21},
				Value: []byte{byte(reason)},
			},
		},
	}

	ca.RevokedCerts[serialNumber] = &revokedCert

	log.Printf(" revoked certificate for CA: %s, Serial: %s", caName, serialNumber)

	return CertificateResponse{
		Success: true,
		Message: "证书撤销成功",
	}
}

func (manager *CAManager) FindCertIssuer(serialNumber string) (*CA, *x509.Certificate, error) {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()

	for _, ca := range manager.CAs {
		if cert, exists := ca.IssuedCerts[serialNumber]; exists {
			return ca, cert, nil
		}
	}
	return nil, nil, fmt.Errorf("the certificate's (Serial: %s) CA is not found", serialNumber)
}

func (manager *CAManager) VerifyXORResult(xorResult []byte, subjectInfo pkix.Name, remainders []*big.Int) bool {
	subjectInfoBytes, err := json.Marshal(subjectInfo)
	if err != nil {
		log.Fatalf("序列化主体信息时出错: %v", err)
	}

	for _, remainder := range remainders {
		remainderBytes := remainder.Bytes()

		for i := 0; i < len(xorResult); i++ {
			xorResult[i] = xorResult[i] ^ remainderBytes[i%len(remainderBytes)]
		}
	}
	if bytes.Equal(xorResult, subjectInfoBytes) {
		log.Printf("true")
		return true
	} else {
		log.Printf("false")
		return false
	}
}
