package cer_ca_tools

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

// CertGenerator 证书生成器
type CertGenerator struct {
	CertsDir string
}

// NewCertGenerator 创建新的证书生成器
func NewCertGenerator(CertsDir string) *CertGenerator {
	return &CertGenerator{
		CertsDir: CertsDir,
	}
}

// GenerateCA 生成CA证书和私钥
func (cg *CertGenerator) GenerateCA() error {

	// 生成椭圆曲线私钥
	caPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("生成CA椭圆曲线私钥失败: %v", err)
	}

	//生成随机序列号
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 160)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)

	// 创建CA证书模板
	caTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"Beijing"},
			Locality:           []string{"Beijing"},
			Organization:       []string{"Test CA"},
			OrganizationalUnit: []string{"IT"},
			CommonName:         "Test Root CA (ECDSA)", //Test Root CA (ECDSA), ca_test_one, ca_test_two, ca_test_three
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour), // 10年
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2, // 允许2级证书链
		MaxPathLenZero:        false,
	}

	// 生成CA证书
	caCertDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return fmt.Errorf("生成CA证书失败: %v", err)
	}

	// 保存CA证书
	caCertPath := filepath.Join(cg.CertsDir, "ca.crt")
	caCertFile, err := os.Create(caCertPath)
	if err != nil {
		return fmt.Errorf("创建CA证书文件失败: %v", err)
	}
	defer caCertFile.Close()

	err = pem.Encode(caCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caCertDER,
	})
	if err != nil {
		return fmt.Errorf("编码CA证书失败: %v", err)
	}

	// 保存CA私钥
	caKeyPath := filepath.Join(cg.CertsDir, "ca.key")
	caKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		return fmt.Errorf("创建CA私钥文件失败: %v", err)
	}
	defer caKeyFile.Close()

	caPrivKeyDER, err := x509.MarshalPKCS8PrivateKey(caPrivKey)
	if err != nil {
		return fmt.Errorf("序列化CA私钥失败: %v", err)
	}

	err = pem.Encode(caKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: caPrivKeyDER,
	})
	if err != nil {
		return fmt.Errorf("编码CA私钥失败: %v", err)
	}

	log.Printf("CA证书已生成（椭圆曲线 P384）: %s", caCertPath)
	log.Printf("CA私钥已生成: %s", caKeyPath)
	return nil
}

func (cg *CertGenerator) GenerateVerifierCert() error {

	log.Println("Generating Server Certificate")

	caCert, caSK, err := cg.LoadCA("ca")
	if err != nil {
		return fmt.Errorf("Error loading CA: %s", err)
	}

	serverSK, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("Error generating server private key: %s", err)
	}

	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Locality:           []string{"xi'an"},
			Province:           []string{"xi'an"},
			Organization:       []string{"Test Server"},
			OrganizationalUnit: []string{"IT"},
			CommonName:         "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:              []string{"localhost", "*.localhost", "127.0.0.1"},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, caCert, &serverSK.PublicKey, caSK)
	if err != nil {
		return fmt.Errorf("生成验证者证书失败: %s", err)
	}

	serverCertPath := filepath.Join(cg.CertsDir, "verifier.crt")
	serverCertFile, err := os.Create(serverCertPath)
	if err != nil {
		return fmt.Errorf("生成验证者证书文件失败: %s", err)
	}
	defer serverCertFile.Close()

	err = pem.Encode(serverCertFile, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertDER,
	})
	if err != nil {
		return fmt.Errorf("编码验证者证书失败: %s", err)
	}

	ServerKeyPath := filepath.Join(cg.CertsDir, "verifier.key")
	ServerKeyFile, err := os.Create(ServerKeyPath)
	if err != nil {
		return fmt.Errorf("生成验证者私钥失败: %s", err)
	}
	defer ServerKeyFile.Close()

	serverSKDER, err := x509.MarshalPKCS8PrivateKey(serverSK)
	if err != nil {
		return fmt.Errorf("序列化验证者密钥失败: %s", err)
	}

	err = pem.Encode(ServerKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: serverSKDER,
	})
	if err != nil {
		return fmt.Errorf("编码成验证者密钥失败: %s", err)
	}

	log.Printf("证书验证者私钥已生成: %s", ServerKeyPath)
	return nil

}

// GenerateClientCert 生成客户端证书
func (cg *CertGenerator) GenerateSubjectCert() error {

	subPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return fmt.Errorf("生成证书主体椭圆曲线私钥失败: %v", err)
	}

	clientKeyPath := filepath.Join(cg.CertsDir, "subject.key")
	clientKeyFile, err := os.Create(clientKeyPath)
	if err != nil {
		return fmt.Errorf("创建证书主体私钥文件失败: %v", err)
	}
	defer clientKeyFile.Close()

	clientPrivKeyDER, err := x509.MarshalPKCS8PrivateKey(subPrivKey)
	if err != nil {
		return fmt.Errorf("序列化证书主体私钥失败: %v", err)
	}

	err = pem.Encode(clientKeyFile, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: clientPrivKeyDER,
	})
	if err != nil {
		return fmt.Errorf("编码证书主体私钥失败: %v", err)
	}

	log.Printf("证书主体私钥已生成: %s", clientKeyPath)
	return nil

	// 加载CA证书和私钥
	//caCert, caPrivKey, err := cg.LoadCA("ca")
	//if err != nil {
	//	return fmt.Errorf("加载CA失败: %v", err)
	//}
	//
	//// 生成客户端椭圆曲线私钥 (P-256适合客户端使用)
	//verPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	//if err != nil {
	//	return fmt.Errorf("生成客户端椭圆曲线私钥失败: %v", err)
	//}
	//
	//// 创建客户端证书模板
	//clientTemplate := x509.Certificate{
	//	SerialNumber: big.NewInt(3),
	//	Subject: pkix.Name{
	//		Country:            []string{"CN"},
	//		Province:           []string{"Beijing"},
	//		Locality:           []string{"Beijing"},
	//		Organization:       []string{"Test Client"},
	//		OrganizationalUnit: []string{"IT"},
	//		CommonName:         "Test Verifier",
	//	},
	//	NotBefore:   time.Now(),
	//	NotAfter:    time.Now().Add(365 * 24 * time.Hour), // 1年
	//	KeyUsage:    x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
	//	ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	//}
	//
	//// 确保CA私钥是正确的类型
	//caPrivateKey, ok := caPrivKey.(*ecdsa.PrivateKey)
	//if !ok {
	//	return fmt.Errorf("CA私钥类型错误，期望*ecdsa.PrivateKey，得到%T", caPrivKey)
	//}
	//
	//// 生成客户端证书（确保公钥类型正确）
	//clientCertDER, err := x509.CreateCertificate(rand.Reader, &clientTemplate, caCert, &verPrivKey.PublicKey, caPrivateKey)
	//if err != nil {
	//	return fmt.Errorf("生成客户端证书失败: %v", err)
	//}
	//
	//// 保存客户端证书
	//clientCertPath := filepath.Join(cg.CertsDir, "verifier.crt")
	//clientCertFile, err := os.Create(clientCertPath)
	//if err != nil {
	//	return fmt.Errorf("创建客户端证书文件失败: %v", err)
	//}
	//defer clientCertFile.Close()
	//
	//err = pem.Encode(clientCertFile, &pem.Block{
	//	Type:  "CERTIFICATE",
	//	Bytes: clientCertDER,
	//})
	//if err != nil {
	//	return fmt.Errorf("编码客户端证书失败: %v", err)
	//}
	//
	//// 保存客户端私钥
	//clientKeyPath := filepath.Join(cg.CertsDir, "verifier.key")
	//clientKeyFile, err := os.Create(clientKeyPath)
	//if err != nil {
	//	return fmt.Errorf("创建客户端私钥文件失败: %v", err)
	//}
	//defer clientKeyFile.Close()
	//
	//clientPrivKeyDER, err := x509.MarshalPKCS8PrivateKey(verPrivKey)
	//if err != nil {
	//	return fmt.Errorf("序列化客户端私钥失败: %v", err)
	//}
	//
	//err = pem.Encode(clientKeyFile, &pem.Block{
	//	Type:  "PRIVATE KEY",
	//	Bytes: clientPrivKeyDER,
	//})
	//if err != nil {
	//	return fmt.Errorf("编码客户端私钥失败: %v", err)
	//}
	//
	//log.Printf("客户端证书已生成（椭圆曲线 P-256）: %s", clientCertPath)
	//log.Printf("客户端私钥已生成: %s", clientKeyPath)
	//return nil
}

func (cg *CertGenerator) LoadCA(caName string) (*x509.Certificate, interface{}, error) {
	// 读取CA证书
	caCertPath := filepath.Join(cg.CertsDir, caName+".crt")
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取CA证书失败: %v", err)
	}

	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return nil, nil, fmt.Errorf("解码CA证书失败")
	}

	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("解析CA证书失败: %v", err)
	}

	// 读取CA私钥
	caKeyPath := filepath.Join(cg.CertsDir, caName+".key")
	caKeyPEM, err := os.ReadFile(caKeyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取CA私钥失败: %v", err)
	}

	caKeyBlock, _ := pem.Decode(caKeyPEM)
	if caKeyBlock == nil {
		return nil, nil, fmt.Errorf("解码CA私钥失败")
	}

	caPrivKey, err := x509.ParsePKCS8PrivateKey(caKeyBlock.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("解析CA私钥失败: %v", err)
	}

	return caCert, caPrivKey, nil
}

// GenerateAllCerts 生成所有证书
func (cg *CertGenerator) GenerateAllCerts() error {
	// 创建证书目录
	err := os.MkdirAll(cg.CertsDir, 0755)
	if err != nil {
		return fmt.Errorf("创建证书目录失败: %v", err)
	}

	// 生成CA证书
	err = cg.GenerateCA()
	if err != nil {
		return err
	}

	// 生成服务器证书(验证者)
	err = cg.GenerateVerifierCert()
	if err != nil {
		return err
	}

	// 生成客户端证书（证书主体）
	err = cg.GenerateSubjectCert()
	if err != nil {
		return err
	}

	log.Println("所有证书生成完成!")
	return nil
}

// ValidateCertChain 验证证书链，模拟验证者行为
func (cg *CertGenerator) ValidateCertChain(certPath string, caCertPath string) error {
	// 读取待验证的证书
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取证书 '%s' 失败: %v", certPath, err)
	}
	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("解码证书 '%s' 失败", certPath)
	}
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书 '%s' 失败: %v", certPath, err)
	}

	// 读取CA证书
	caCertPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("读取CA证书 '%s' 失败: %v", caCertPath, err)
	}
	caCertBlock, _ := pem.Decode(caCertPEM)
	if caCertBlock == nil {
		return fmt.Errorf("解码CA证书 '%s' 失败", caCertPath)
	}
	caCert, err := x509.ParseCertificate(caCertBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析CA证书 '%s' 失败: %v", caCertPath, err)
	}

	// 创建一个证书池，并添加CA证书
	roots := x509.NewCertPool()
	roots.AddCert(caCert)

	// 设置验证选项
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: x509.NewCertPool(), // 如果有中间CA，也需要添加
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	// 验证证书链
	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("证书链验证失败 for '%s': %v", certPath, err)
	}

	// 额外检查有效期
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("证书 '%s' 尚未生效", certPath)
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("证书 '%s' 已过期", certPath)
	}

	log.Printf(" 证书链验证成功: '%s' 由 '%s' 有效签发.", filepath.Base(certPath), caCert.Subject.CommonName)
	log.Printf("  主体: %s", cert.Subject)
	log.Printf("  有效期: %s 到 %s", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02"))
	return nil
}

// ValidateCertificate 验证单个证书的基本信息
func (cg *CertGenerator) ValidateCertificate(certPath string) error {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("读取证书文件失败: %v", err)
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return fmt.Errorf("解码证书失败")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("解析证书失败: %v", err)
	}

	// 验证证书有效期
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("证书尚未生效")
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("证书已过期")
	}

	log.Printf(" 单个证书验证成功: %s", filepath.Base(certPath))
	log.Printf("  主体: %s", cert.Subject)
	log.Printf("  颁发者: %s", cert.Issuer)
	log.Printf("  有效期: %s 到 %s", cert.NotBefore.Format("2006-01-02 15:04:05"), cert.NotAfter.Format("2006-01-02 15:04:05"))
	log.Printf("  序列号: %s", cert.SerialNumber)
	log.Printf("  是否为CA: %t", cert.IsCA)

	return nil
}
