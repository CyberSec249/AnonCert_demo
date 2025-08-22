package main

import (
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/cer_ca_tools"
	"log"
	"net/http"
	"time"

	// "net/http"
	"os"
	"path/filepath"
)

func main() {
	//初始化CA、证书主体、验证者密钥和证书
	//currentDir, _ := os.Getwd()
	//
	//certsDir := filepath.Join(currentDir, "certs")
	//
	//generator := cer_ca_tools.NewCertGenerator(certsDir)
	//
	//keyAndCertGenerate(generator)

	//设置CA HTTP接口
	//setupCAHTTP()

	//bloom Filter
	BCCBFSet()

}

func keyAndCertGenerate(cg *cer_ca_tools.CertGenerator) {

	if err := os.MkdirAll(cg.CertsDir, 0755); err != nil {
		log.Fatalf("创建证书目录 '%s' 失败: %v", cg.CertsDir, err)
	}

	err := cg.GenerateAllCerts()
	if err != nil {
		log.Fatalf("<UNK> '%s' <UNK>: %v", cg.CertsDir, err)
	}

	caCertPath := filepath.Join(cg.CertsDir, "ca.crt")
	verifierCertPath := filepath.Join(cg.CertsDir, "verifier.crt")

	log.Println("\n[验证] 验证CA证书本身 (自签名)...")
	if err := cg.ValidateCertificate(caCertPath); err != nil {
		log.Printf("CA证书验证失败: %v", err)
	}

	log.Println("\n[验证] 验证客户端证书链...")
	if err := cg.ValidateCertChain(verifierCertPath, caCertPath); err != nil {
		log.Printf("客户端证书链验证失败: %v", err)
	}

}

func setupCAHTTP() {

	caManager := cer_ca_tools.NewCAManager()

	primePool := cer_ca_tools.NewPrimePool()

	// 生成几个大质数用于演示
	err := primePool.GeneratePrimes(1000, 64) // 生成5个64位的质数
	if err != nil {
		log.Fatalf("生成质数时出错: %v", err)
	}

	caManager.PrimePool = primePool

	caConfigs := []struct{ caName string }{{"ca_test_one"}, {"ca_test_two"}, {"ca_test_three"}}

	for _, caConfig := range caConfigs {
		ca, err := cer_ca_tools.CreateNewCA(caConfig.caName)
		if err != nil {
			log.Fatal("creat CA failed", caConfig.caName, err)
		}
		caManager.AddCAToManager(ca)
	}

	caManager.SetupHTTPHandlers()

	log.Println(" Starting HTTP server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

	// 阻塞主程序，防止退出
	select {}

}

func BCCBFSet() {
	bloomFilter := cer_ca_tools.NewCountingBloomFilter(100, 0.01, 1)

	//revokedSerials := []string{
	//	"00000000000000000",
	//	"00000000000000001",
	//	"00000000000000002",
	//	"00000000000000003",
	//}

	var revokedSerials []string
	for i := 0; i < 10000; i++ {
		serial := fmt.Sprintf("%017d", i)
		revokedSerials = append(revokedSerials, serial)
	}

	for _, serial := range revokedSerials {
		bloomFilter.AddElement([]byte(serial))
	}

	//bloomFilter.PrintStats()
	queryData := "00000000000000020"
	startBCCBFQuery := time.Now()

	if !bloomFilter.QueryElement([]byte(queryData)) {
		fmt.Println("元素", queryData, "不存在")
	} else {
		fmt.Println("元素", queryData, "可能存在")
	}

	endBCCBFQuery := time.Since(startBCCBFQuery)
	fmt.Println("BCCBF Query Time:", endBCCBFQuery)
}
