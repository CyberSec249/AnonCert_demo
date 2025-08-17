package main

//
//import (
//	"crypto/ecdsa"
//	"crypto/elliptic"
//	"crypto/rand"
//	"crypto/x509/pkix"
//	"fmt"
//	"github.com/FISCO-BCOS/go-sdk/cer_ca_tools"
//	"github.com/FISCO-BCOS/go-sdk/cer_subject_tools"
//	"github.com/FISCO-BCOS/go-sdk/cert_vrf"
//	"log"
//	"sync"
//	"time"
//)
//
//// PerformanceResults 存储性能测试结果
//type PerformanceResults struct {
//	OperationType    string        `json:"operation_type"`
//	TotalTime        time.Duration `json:"total_time"`
//	AverageTime      time.Duration `json:"average_time"`
//	MinTime          time.Duration `json:"min_time"`
//	MaxTime          time.Duration `json:"max_time"`
//	OperationCount   int           `json:"operation_count"`
//	ThroughputPerSec float64       `json:"throughput_per_sec"`
//}
//
//// PerformanceTester 性能测试器
//type PerformanceTester struct {
//	caManager  *cer_ca_tools.CAManager
//	subjects   []*cer_subject_tools.Subject
//	vrfManager *cert_vrf.VRFManager
//	results    []PerformanceResults
//	mu         sync.Mutex
//}
//
//func main() {
//	log.Println("Starting AnonCert Performance Tests...")
//
//	tester := NewPerformanceTester()
//	if err := tester.RunAllTests(); err != nil {
//		log.Fatalf("Performance tests failed: %v", err)
//	}
//
//	log.Println("Performance tests completed successfully!")
//}
//
//// NewPerformanceTester 创建新的性能测试器
//func NewPerformanceTester() *PerformanceTester {
//	return &PerformanceTester{
//		caManager:  cer_ca_tools.NewCAManager(),
//		vrfManager: cert_vrf.NewVRFManager(),
//		results:    make([]PerformanceResults, 0),
//	}
//}
//
//// SetupTestEnvironment 设置测试环境
//func (pt *PerformanceTester) SetupTestEnvironment() error {
//	log.Println("Setting up test environment...")
//
//	// 创建多个CA用于测试
//	caNames := []string{"ca_test_one", "ca_test_two", "ca_test_three"}
//	for _, caName := range caNames {
//		ca, err := cer_ca_tools.CreateNewCA(caName)
//		if err != nil {
//			return fmt.Errorf("failed to create CA %s: %v", caName, err)
//		}
//		pt.caManager.AddCAToManager(ca)
//		log.Printf("Created CA: %s", caName)
//	}
//
//	// 创建多个Subject用于测试
//	for i := 0; i < 10; i++ {
//		subject := cer_subject_tools.NewSubject("http://localhost:8080")
//
//		// 生成密钥对
//		privateKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
//		if err != nil {
//			return fmt.Errorf("failed to generate key pair for subject %d: %v", i, err)
//		}
//		subject.PrivateKey = privateKey
//		subject.PublicKey = &privateKey.PublicKey
//
//		pt.subjects = append(pt.subjects, subject)
//	}
//
//	log.Printf("Created %d subjects for testing", len(pt.subjects))
//	return nil
//}
//
//// TestCertificateIssuance 测试证书签发性能（包含CRT参数生成和验证）
//func (pt *PerformanceTester) TestCertificateIssuance(iterations int) {
//	log.Printf("Testing certificate issuance performance with %d iterations...", iterations)
//
//	var crtGenTimes []time.Duration
//	var crtVerifyTimes []time.Duration
//	var certIssueTimes []time.Duration
//	var totalCrtGenTime, totalCrtVerifyTime, totalCertIssueTime time.Duration
//	var minCrtGenTime, maxCrtGenTime, minCrtVerifyTime, maxCrtVerifyTime time.Duration
//	var minCertIssueTime, maxCertIssueTime time.Duration
//
//	caName := "ca_test_one"
//	ca, exists := pt.caManager.GetCAInfo(caName)
//	if !exists {
//		log.Printf("CA %s not found", caName)
//		return
//	}
//
//	for i := 0; i < iterations; i++ {
//		// 创建测试用的主体信息
//		subjectInfo := pkix.Name{
//			CommonName:         fmt.Sprintf("test-subject-%d", i),
//			Organization:       []string{"Test Org"},
//			OrganizationalUnit: []string{"Test Unit"},
//			Country:            []string{"CN"},
//			Province:           []string{"Test Province"},
//			Locality:           []string{"Test Locality"},
//		}
//
//		// 生成临时密钥对
//		privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
//		publicKey := &privateKey.PublicKey
//
//		// 1. 测量CRT参数生成时间（证书主体侧）
//		startCrtGen := time.Now()
//		crtOps, xorResult, remainders := pt.performCRTGeneration(subjectInfo)
//		crtGenDuration := time.Since(startCrtGen)
//
//		if crtOps == nil {
//			log.Printf("CRT generation failed for iteration %d", i)
//			continue
//		}
//
//		// 2. 测量CA验证CRT参数时间
//		startCrtVerify := time.Now()
//		isValid := pt.caManager.VerifyXORResult(xorResult, subjectInfo, remainders)
//		crtVerifyDuration := time.Since(startCrtVerify)
//
//		if !isValid {
//			log.Printf("CRT verification failed for iteration %d", i)
//			continue
//		}
//
//		// 3. 测量证书签发时间（CA侧）
//		startCertIssue := time.Now()
//		response := ca.IssueCertificate(subjectInfo, publicKey)
//		certIssueDuration := time.Since(startCertIssue)
//
//		if !response.Success {
//			log.Printf("Certificate issuance failed: %s", response.Message)
//			continue
//		}
//
//		// 记录时间
//		crtGenTimes = append(crtGenTimes, crtGenDuration)
//		crtVerifyTimes = append(crtVerifyTimes, crtVerifyDuration)
//		certIssueTimes = append(certIssueTimes, certIssueDuration)
//
//		totalCrtGenTime += crtGenDuration
//		totalCrtVerifyTime += crtVerifyDuration
//		totalCertIssueTime += certIssueDuration
//
//		// 更新最小最大时间
//		if i == 0 {
//			minCrtGenTime, maxCrtGenTime = crtGenDuration, crtGenDuration
//			minCrtVerifyTime, maxCrtVerifyTime = crtVerifyDuration, crtVerifyDuration
//			minCertIssueTime, maxCertIssueTime = certIssueDuration, certIssueDuration
//		} else {
//			if crtGenDuration < minCrtGenTime {
//				minCrtGenTime = crtGenDuration
//			}
//			if crtGenDuration > maxCrtGenTime {
//				maxCrtGenTime = crtGenDuration
//			}
//			if crtVerifyDuration < minCrtVerifyTime {
//				minCrtVerifyTime = crtVerifyDuration
//			}
//			if crtVerifyDuration > maxCrtVerifyTime {
//				maxCrtVerifyTime = crtVerifyDuration
//			}
//			if certIssueDuration < minCertIssueTime {
//				minCertIssueTime = certIssueDuration
//			}
//			if certIssueDuration > maxCertIssueTime {
//				maxCertIssueTime = certIssueDuration
//			}
//		}
//	}
//
//	if len(times) > 0 {
//		avgTime := totalTime / time.Duration(len(times))
//		throughput := float64(len(times)) / totalTime.Seconds()
//
//		result := PerformanceResults{
//			OperationType:    "Certificate Issuance",
//			TotalTime:        totalTime,
//			AverageTime:      avgTime,
//			MinTime:          minTime,
//			MaxTime:          maxTime,
//			OperationCount:   len(times),
//			ThroughputPerSec: throughput,
//		}
//
//		pt.mu.Lock()
//		pt.results = append(pt.results, result)
//		pt.mu.Unlock()
//
//		log.Printf("Certificate Issuance Results:")
//		log.Printf("  Total Operations: %d", len(times))
//		log.Printf("  Total Time: %v", totalTime)
//		log.Printf("  Average Time: %v", avgTime)
//		log.Printf("  Min Time: %v", minTime)
//		log.Printf("  Max Time: %v", maxTime)
//		log.Printf("  Throughput: %.2f ops/sec", throughput)
//	}
//}
//
//// TestCertificateRevocation 测试证书撤销性能
//func (pt *PerformanceTester) TestCertificateRevocation(iterations int) {
//	log.Printf("Testing certificate revocation performance with %d iterations...", iterations)
//
//	caName := "ca_test_one"
//	ca, exists := pt.caManager.GetCAInfo(caName)
//	if !exists {
//		log.Printf("CA %s not found", caName)
//		return
//	}
//
//	// 首先签发一些证书用于撤销测试
//	var issuedCerts []string
//	for i := 0; i < iterations; i++ {
//		subjectInfo := pkix.Name{
//			CommonName:   fmt.Sprintf("revoke-test-subject-%d", i),
//			Organization: []string{"Revoke Test Org"},
//		}
//
//		privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
//		publicKey := &privateKey.PublicKey
//
//		response := ca.IssueCertificate(subjectInfo, publicKey)
//		if response.Success {
//			// 从CA的IssuedCerts中获取最新的序列号
//			for serialNum := range ca.IssuedCerts {
//				issuedCerts = append(issuedCerts, serialNum)
//				break // 只取最新的一个
//			}
//		}
//	}
//
//	var times []time.Duration
//	var totalTime time.Duration
//	var minTime, maxTime time.Duration
//
//	for i, serialNum := range issuedCerts {
//		startTime := time.Now()
//		response := ca.RevokeCertificate(caName, serialNum, 1) // reason: 1 = key compromise
//		duration := time.Since(startTime)
//
//		if !response.Success {
//			log.Printf("Certificate revocation failed: %s", response.Message)
//			continue
//		}
//
//		times = append(times, duration)
//		totalTime += duration
//
//		if i == 0 || duration < minTime {
//			minTime = duration
//		}
//		if i == 0 || duration > maxTime {
//			maxTime = duration
//		}
//	}
//
//	if len(times) > 0 {
//		avgTime := totalTime / time.Duration(len(times))
//		throughput := float64(len(times)) / totalTime.Seconds()
//
//		result := PerformanceResults{
//			OperationType:    "Certificate Revocation",
//			TotalTime:        totalTime,
//			AverageTime:      avgTime,
//			MinTime:          minTime,
//			MaxTime:          maxTime,
//			OperationCount:   len(times),
//			ThroughputPerSec: throughput,
//		}
//
//		pt.mu.Lock()
//		pt.results = append(pt.results, result)
//		pt.mu.Unlock()
//
//		log.Printf("Certificate Revocation Results:")
//		log.Printf("  Total Operations: %d", len(times))
//		log.Printf("  Total Time: %v", totalTime)
//		log.Printf("  Average Time: %v", avgTime)
//		log.Printf("  Min Time: %v", minTime)
//		log.Printf("  Max Time: %v", maxTime)
//		log.Printf("  Throughput: %.2f ops/sec", throughput)
//	}
//}
//
//// TestVRFVerification 测试VRF验证性能
//func (pt *PerformanceTester) TestVRFVerification(iterations int) {
//	log.Printf("Testing VRF verification performance with %d iterations...", iterations)
//
//	var times []time.Duration
//	var totalTime time.Duration
//	var minTime, maxTime time.Duration
//
//	for i := 0; i < iterations; i++ {
//		sessionID := fmt.Sprintf("test-session-%d", i)
//
//		// 生成挑战
//		challenge, err := pt.vrfManager.GenerateVRFChallenge(sessionID)
//		if err != nil {
//			log.Printf("Failed to generate challenge: %v", err)
//			continue
//		}
//
//		// 使用测试主体的私钥生成VRF证明
//		subject := pt.subjects[i%len(pt.subjects)]
//
//		vrfKeyPair := &cert_vrf.VRFKeyPair{
//			PublicKey:  subject.PublicKey,
//			PrivateKey: subject.PrivateKey,
//		}
//
//		proof, err := pt.vrfManager.GenerateVRFProof(vrfKeyPair, challenge)
//		if err != nil {
//			log.Printf("Failed to generate VRF proof: %v", err)
//			continue
//		}
//
//		// 测量VRF验证时间
//		startTime := time.Now()
//		isValid, err := pt.vrfManager.VerifyVRFProof(subject.PublicKey, challenge, proof)
//		duration := time.Since(startTime)
//
//		if !isValid {
//			log.Printf("VRF verification failed for iteration %d", i)
//			continue
//		}
//
//		times = append(times, duration)
//		totalTime += duration
//
//		if i == 0 || duration < minTime {
//			minTime = duration
//		}
//		if i == 0 || duration > maxTime {
//			maxTime = duration
//		}
//	}
//
//	if len(times) > 0 {
//		avgTime := totalTime / time.Duration(len(times))
//		throughput := float64(len(times)) / totalTime.Seconds()
//
//		result := PerformanceResults{
//			OperationType:    "VRF Verification",
//			TotalTime:        totalTime,
//			AverageTime:      avgTime,
//			MinTime:          minTime,
//			MaxTime:          maxTime,
//			OperationCount:   len(times),
//			ThroughputPerSec: throughput,
//		}
//
//		pt.mu.Lock()
//		pt.results = append(pt.results, result)
//		pt.mu.Unlock()
//
//		log.Printf("VRF Verification Results:")
//		log.Printf("  Total Operations: %d", len(times))
//		log.Printf("  Total Time: %v", totalTime)
//		log.Printf("  Average Time: %v", avgTime)
//		log.Printf("  Min Time: %v", minTime)
//		log.Printf("  Max Time: %v", maxTime)
//		log.Printf("  Throughput: %.2f ops/sec", throughput)
//	}
//}
//
//// TestConcurrentOperations 测试并发操作性能
//func (pt *PerformanceTester) TestConcurrentOperations(operationType string, iterations int, concurrency int) {
//	log.Printf("Testing concurrent %s performance with %d iterations and %d goroutines...",
//		operationType, iterations, concurrency)
//
//	var wg sync.WaitGroup
//	var times []time.Duration
//	var mu sync.Mutex
//
//	totalStartTime := time.Now()
//
//	for i := 0; i < concurrency; i++ {
//		wg.Add(1)
//		go func(workerID int) {
//			defer wg.Done()
//
//			iterationsPerWorker := iterations / concurrency
//			if workerID < iterations%concurrency {
//				iterationsPerWorker++
//			}
//
//			for j := 0; j < iterationsPerWorker; j++ {
//				var duration time.Duration
//
//				switch operationType {
//				case "issuance":
//					duration = pt.performSingleIssuance(workerID, j)
//				case "revocation":
//					duration = pt.performSingleRevocation(workerID, j)
//				case "verification":
//					duration = pt.performSingleVerification(workerID, j)
//				}
//
//				if duration > 0 {
//					mu.Lock()
//					times = append(times, duration)
//					mu.Unlock()
//				}
//			}
//		}(i)
//	}
//
//	wg.Wait()
//	totalTime := time.Since(totalStartTime)
//
//	if len(times) > 0 {
//		var totalDuration time.Duration
//		minTime := times[0]
//		maxTime := times[0]
//
//		for _, t := range times {
//			totalDuration += t
//			if t < minTime {
//				minTime = t
//			}
//			if t > maxTime {
//				maxTime = t
//			}
//		}
//
//		avgTime := totalDuration / time.Duration(len(times))
//		throughput := float64(len(times)) / totalTime.Seconds()
//
//		result := PerformanceResults{
//			OperationType:    fmt.Sprintf("Concurrent %s", operationType),
//			TotalTime:        totalTime,
//			AverageTime:      avgTime,
//			MinTime:          minTime,
//			MaxTime:          maxTime,
//			OperationCount:   len(times),
//			ThroughputPerSec: throughput,
//		}
//
//		pt.mu.Lock()
//		pt.results = append(pt.results, result)
//		pt.mu.Unlock()
//
//		log.Printf("Concurrent %s Results:", operationType)
//		log.Printf("  Total Operations: %d", len(times))
//		log.Printf("  Total Time: %v", totalTime)
//		log.Printf("  Average Time: %v", avgTime)
//		log.Printf("  Min Time: %v", minTime)
//		log.Printf("  Max Time: %v", maxTime)
//		log.Printf("  Throughput: %.2f ops/sec", throughput)
//	}
//}
//
//// 辅助方法：执行单个证书签发
//func (pt *PerformanceTester) performSingleIssuance(workerID, iterationID int) time.Duration {
//	caName := "ca_test_one"
//	ca, exists := pt.caManager.GetCAInfo(caName)
//	if !exists {
//		return 0
//	}
//
//	subjectInfo := pkix.Name{
//		CommonName:   fmt.Sprintf("worker-%d-iter-%d", workerID, iterationID),
//		Organization: []string{"Concurrent Test"},
//	}
//
//	privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
//	publicKey := &privateKey.PublicKey
//
//	startTime := time.Now()
//	response := ca.IssueCertificate(subjectInfo, publicKey)
//	duration := time.Since(startTime)
//
//	if !response.Success {
//		return 0
//	}
//
//	return duration
//}
//
//// 辅助方法：执行单个证书撤销（需要先签发）
//func (pt *PerformanceTester) performSingleRevocation(workerID, iterationID int) time.Duration {
//	caName := "ca_test_one"
//	ca, exists := pt.caManager.GetCAInfo(caName)
//	if !exists {
//		return 0
//	}
//
//	// 先签发证书
//	subjectInfo := pkix.Name{
//		CommonName:   fmt.Sprintf("revoke-worker-%d-iter-%d", workerID, iterationID),
//		Organization: []string{"Revoke Test"},
//	}
//
//	privateKey, _ := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
//	publicKey := &privateKey.PublicKey
//
//	response := ca.IssueCertificate(subjectInfo, publicKey)
//	if !response.Success {
//		return 0
//	}
//
//	// 获取序列号（简化处理，实际应该从响应中解析）
//	var serialNum string
//	for serial := range ca.IssuedCerts {
//		serialNum = serial
//		break
//	}
//
//	// 测量撤销时间
//	startTime := time.Now()
//	revokeResponse := ca.RevokeCertificate(caName, serialNum, 1)
//	duration := time.Since(startTime)
//
//	if !revokeResponse.Success {
//		return 0
//	}
//
//	return duration
//}
//
//// 辅助方法：执行单个VRF验证
//func (pt *PerformanceTester) performSingleVerification(workerID, iterationID int) time.Duration {
//	sessionID := fmt.Sprintf("worker-%d-iter-%d", workerID, iterationID)
//
//	challenge, err := pt.vrfManager.GenerateVRFChallenge(sessionID)
//	if err != nil {
//		return 0
//	}
//
//	subject := pt.subjects[workerID%len(pt.subjects)]
//	vrfKeyPair := &cert_vrf.VRFKeyPair{
//		PublicKey:  subject.PublicKey,
//		PrivateKey: subject.PrivateKey,
//	}
//	proof, err := pt.vrfManager.GenerateVRFProof(vrfKeyPair, challenge)
//	if err != nil {
//		return 0
//	}
//
//	startTime := time.Now()
//	isValid, err := pt.vrfManager.VerifyVRFProof(subject.PublicKey, challenge, proof)
//	duration := time.Since(startTime)
//
//	if !isValid {
//		return 0
//	}
//
//	return duration
//}
//
//// PrintResults 打印所有测试结果
//func (pt *PerformanceTester) PrintResults() {
//	log.Println("\n=== Performance Test Results Summary ===")
//	for _, result := range pt.results {
//		log.Printf("\nOperation: %s", result.OperationType)
//		log.Printf("  Total Operations: %d", result.OperationCount)
//		log.Printf("  Total Time: %v", result.TotalTime)
//		log.Printf("  Average Time: %v", result.AverageTime)
//		log.Printf("  Min Time: %v", result.MinTime)
//		log.Printf("  Max Time: %v", result.MaxTime)
//		log.Printf("  Throughput: %.2f ops/sec", result.ThroughputPerSec)
//		log.Printf("  Average Time (ms): %.2f", float64(result.AverageTime.Nanoseconds())/1000000)
//	}
//}
//
//// RunAllTests 运行所有性能测试
//func (pt *PerformanceTester) RunAllTests() error {
//	// 设置测试环境
//	if err := pt.SetupTestEnvironment(); err != nil {
//		return fmt.Errorf("failed to setup test environment: %v", err)
//	}
//
//	// 测试参数
//	iterations := 100
//	concurrency := 10
//
//	// 顺序测试
//	pt.TestCertificateIssuance(iterations)
//	pt.TestCertificateRevocation(iterations)
//	pt.TestVRFVerification(iterations)
//
//	// 并发测试
//	pt.TestConcurrentOperations("issuance", iterations, concurrency)
//	pt.TestConcurrentOperations("verification", iterations, concurrency)
//
//	// 打印结果
//	pt.PrintResults()
//
//	return nil
//}
