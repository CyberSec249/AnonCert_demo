package main

import (
	"database/sql"
	"encoding/json"
	"github.com/FISCO-BCOS/go-sdk/helloworld"
	"github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"time"
)

type Metrics struct {
	NodeCount     string `json:"nodeCount"`
	BlockCount    string `json:"blockCount"`
	TxCount       string `json:"txCount"`
	ContractCount string `json:"contractCount"`
}

type RequestCertInfo struct {
	ID          int64  `json:"id"`
	PublicKey   string `json:"publicKey"`
	CA          string `json:"ca"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/metrics", metricsHandler)
	mux.HandleFunc("/api/request/list", RequestListHandler)
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

/*
统一的 JSON 响应
*/
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
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

/*
/api/metrics 处理器
*/
func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// 这里也可以做缓存：例如每 5s 触发一次真实查询，其余读内存

	blockNum, txNum, nodeNum, contractNum := helloworld.BlockInfoGet()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	resp := Metrics{
		NodeCount:     nodeNum,
		BlockCount:    blockNum,
		TxCount:       txNum,
		ContractCount: contractNum,
	}
	writeJSON(w, http.StatusOK, resp)
}

/*
/api/request/list 处理器
*/
func RequestListHandler(w http.ResponseWriter, r *http.Request) {
	cfg := mysql.NewConfig()
	cfg.User = "dpki_user"
	cfg.Passwd = "bc@xdu308"
	cfg.Net = "tcp"
	cfg.Addr = "127.0.0.1:3306"
	cfg.DBName = "dpki"
	cfg.Params = map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "true",
		"loc":       "Local",
	}
	cfg.AllowNativePasswords = true

	dsn := cfg.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		http.Error(w, "ping db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// ✅ 明确列顺序，只取前端需要的 5 列
	rows, err := db.Query(`SELECT ID, PublicKey, CA, Status, Description FROM requestCertList`)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []RequestCertInfo
	for rows.Next() {
		var item RequestCertInfo
		if err := rows.Scan(&item.ID, &item.PublicKey, &item.CA, &item.Status, &item.Description); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"items": list,
	})
}

//configs, err := conf.ParseConfigFile("config.toml")
//if err != nil {
//	log.Fatal(err)
//}
//
//config := &configs[0]
//
//c, err := client.Dial(config)
//if err != nil {
//	log.Fatal(err)
//}
//
//address, tx, instance, err := helloworld.DeployHelloWorld(c.GetTransactOpts(), c) // deploy contract
//if err != nil {
//	log.Fatal(err)
//}
//
//fmt.Println("HelloWorld contract address: ", address.Hex()) // the address should be saved
//fmt.Println("transaction hash: ", tx.Hash().Hex())
//
//_ = instance
