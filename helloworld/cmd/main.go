package main

import (
	"encoding/json"
	"fmt"
	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	hellowrold "github.com/FISCO-BCOS/go-sdk/helloworld"
	contractGo "github.com/FISCO-BCOS/go-sdk/helloworld/contractFile"
	"log"
	"net/http"
	"time"
)

var primePool *hellowrold.PrimePool

func main() {
	//deployContract()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/ca/list", hellowrold.CAListHandler)
	mux.HandleFunc("/api/ca/register", hellowrold.GenerateCAHandle)
	mux.HandleFunc("/api/ca/cert/download", hellowrold.DownloadCACertHandle())
	mux.HandleFunc("/api/cert/list", hellowrold.CertListHandler)
	mux.HandleFunc("/api/cert/download", hellowrold.DownloadSubjectCertHandle())

	mux.HandleFunc("/api/metrics", hellowrold.MetricsHandler)
	mux.HandleFunc("/api/request/list", hellowrold.RequestListHandler)
	mux.HandleFunc("/api/request/cert", hellowrold.RequestCertHandler)
	mux.HandleFunc("/api/request/crt", RequestCRTHandler)
	mux.HandleFunc("/api/request/detail", hellowrold.GetCertRequestDetail)
	mux.HandleFunc("/api/request/crt/submit", hellowrold.SubmitCRTHandler)
	mux.HandleFunc("/api/request/crt/query", hellowrold.QueryCRTHandler)

	mux.HandleFunc("/api/issue/list", hellowrold.IssueListHandler)
	mux.HandleFunc("/api/issue/subject/detail", hellowrold.IssueSubjectInfoHandler)
	mux.HandleFunc("/api/issue/cert/check", hellowrold.IssueSubjectCheckHandler)
	mux.HandleFunc("/api/issue/cert/reject", hellowrold.IssueSubjectRejectHandler)
	mux.HandleFunc("/api/issue/crt/detail", hellowrold.IssueCRTQueryHandler)
	mux.HandleFunc("/api/issue/crt/verify", hellowrold.IssueCRTVerifyHandler)
	mux.HandleFunc("/api/issue/cert/issuance", hellowrold.IssueCertIssuanceHandler)

	primePool = hellowrold.NewPrimePool()
	err := primePool.GeneratePrimes(1000, 64) // 生成5个64位的质数
	if err != nil {
		log.Fatalf("生成质数时出错: %v", err)
	}

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("DPKI backend listening on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}

}

// /api/request/crt
func RequestCRTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}

	var crtParameter hellowrold.CRTOperations
	var err error

	if err := json.NewDecoder(r.Body).Decode(&crtParameter); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	crtParameter.Moduli, err = primePool.RandomModuli(3)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
	}
	crtParameter.GenerateRandomRemainders()

	crtParameter.SolveChineseRemainderTheorem()

	crtParameterString := hellowrold.ConvertToStringVersion(crtParameter)

	writeJSON(w, http.StatusOK, crtParameterString)
}

/*
简单 CORS 中间件：允许前端本地或部署站点跨域访问
如需更严格，可把 "*" 改为具体域名，例如 http://localhost:5173
*/
func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*" // 非浏览器或本地脚本
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func deployContract() {
	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}

	config := &configs[0]

	c, err := client.Dial(config)
	if err != nil {
		log.Fatal(err)
	}

	address, tx, instance, err := contractGo.DeployCertOperKV(c.GetTransactOpts(), c) // deploy contract
	if err != nil {
		log.Fatal(err)
	}
	//HelloWorld contract address:  0x0a68F060B46e0d8f969383D260c34105EA13a9dd
	//transaction hash:  0xc16cf8f32c15bd8130010400a3e45a89a704727b86da883fa442f342dfc68574
	fmt.Println("HelloWorld contract address: ", address.Hex()) // the address should be saved
	fmt.Println("transaction hash: ", tx.Hash().Hex())

	_ = instance

}
