package helloworld

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	NodeCount     string `json:"nodeCount"`
	BlockCount    string `json:"blockCount"`
	TxCount       string `json:"txCount"`
	ContractCount string `json:"contractCount"`
}

type RequestCertListInfo struct {
	ID          int64  `json:"id"`
	PublicKey   string `json:"publicKey"`
	CAID        string `json:"ca_id"`
	CAPublicKey string `json:"caPublicKey"`
	Status      string `json:"status"`
	Description string `json:"description"`
	SubjectInfo string `json:"subjectInfo"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

/************* 处理器 *************/
// /api/metrics
func MetricsHandler(w http.ResponseWriter, r *http.Request) {
	blockNum, txNum, nodeNum, contractNum := BlockInfoGet()
	resp := Metrics{
		NodeCount:     nodeNum,
		BlockCount:    blockNum,
		TxCount:       txNum,
		ContractCount: contractNum,
	}
	writeJSON(w, http.StatusOK, resp)
}

// /api/request/list
func RequestListHandler(w http.ResponseWriter, r *http.Request) {
	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	const q = `SELECT r.ID, r.PublicKey, r.ca_id, IFNULL(c.caPublicKey, '') AS caPublicKey, r.Status, r.Description, r.SubjectInfo 
		FROM dpki.requestCertList AS r LEFT JOIN dpki.CAList AS c ON c.ca_id = r.ca_id ORDER BY r.ID DESC;`
	rows, err := db.Query(q)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	list := make([]RequestCertListInfo, 0, 64)
	for rows.Next() {
		var item RequestCertListInfo
		if err := rows.Scan(
			&item.ID,
			&item.PublicKey,
			&item.CAID,
			&item.CAPublicKey,
			&item.Status,
			&item.Description,
			&item.SubjectInfo,
		); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

// /api/request/cert
func RequestCertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "invalid form: "+err.Error(), http.StatusBadRequest)
		return
	}
	publicKey := strings.TrimSpace(r.FormValue("publicKey"))
	subjectInfo := strings.TrimSpace(r.FormValue("subjectInfo"))
	caPublicKey := strings.TrimSpace(r.FormValue("caPublicKey"))
	description := strings.TrimSpace(r.FormValue("description"))
	fmt.Println(publicKey, caPublicKey, description)
	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var caID int64
	err = db.QueryRow(`SELECT ca_id FROM dpki.CAList WHERE caPublicKey = ? LIMIT 1`, caPublicKey).Scan(&caID)
	if err == sql.ErrNoRows {
		http.Error(w, "CA不存在: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(caID)

	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "start db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO requestCertList (PublicKey, ca_id, Status, Description, SubjectInfo)
         VALUES (?, ?, 'pending', ?, ?)`,
		publicKey, caID, description, subjectInfo,
	)
	if err != nil {
		http.Error(w, "数据库插入失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	request_id, _ := res.LastInsertId()
	fmt.Println("request_id", request_id)

	if _, err := tx.Exec(
		`INSERT INTO issueCertList (request_id, subject_status, crt_status)
         VALUES (?, '0', '0')`, request_id,
	); err != nil {
		http.Error(w, "数据库插入失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "数据库提交失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"msg":        "ok",
		"request_id": request_id,
		"ca_id":      caID,
		"ts":         time.Now().Format(time.RFC3339),
	})
}

// /api/request/detail
func GetCertRequestDetail(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var item RequestCertListInfo
	err = db.QueryRow(`SELECT r.ID, r.PublicKey, r.ca_id, IFNULL(c.caPublicKey, '') AS caPublicKey, r.Status, r.Description, r.SubjectInfo 
		FROM dpki.requestCertList AS r LEFT JOIN dpki.CAList AS c ON c.ca_id = r.ca_id WHERE id=? LIMIT 1`, id).
		Scan(&item.ID, &item.PublicKey, &item.CAID, &item.CAPublicKey, &item.Status, &item.Description, &item.SubjectInfo)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, item)
}

// /api/request/crt/submit
func SubmitCRTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}

	var crtParameterString CRTOperationsString

	if err := json.NewDecoder(r.Body).Decode(&crtParameterString); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	moduliJSONBytes, err := json.Marshal(crtParameterString.Moduli)
	remaindersJSONBytes, err := json.Marshal(crtParameterString.Remainders)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	//插入CRT数据
	_, err = db.Exec(
		`INSERT INTO requestCertCRT (request_id, moduli, remainders, x)
	    VALUES (?, ?, ?, ?)`,
		crtParameterString.ID, moduliJSONBytes, remaindersJSONBytes, crtParameterString.X,
	)
	if err != nil {
		http.Error(w, "insert error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//更新当前请求状态
	_, err = db.Exec(
		`UPDATE requestCertList SET Status = 'pending' WHERE id =?`, crtParameterString.ID,
	)
	if err != nil {
		http.Error(w, "update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"msg": "crt submit success",
	})
}

// /api/request/crt/query
func QueryCRTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		PublicKey string `json:"publicKey,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var request_id struct {
		ID string `json:"id,omitempty"`
	}
	err = db.QueryRow(`SELECT id FROM requestCertList WHERE PublicKey=?`, in.PublicKey).Scan(&request_id.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	var (
		reqIDStr       string
		moduliJSON     string
		remaindersJSON string
		xStr           string
	)
	err = db.QueryRow(`SELECT request_id, moduli, remainders, x
                       FROM requestCertCRT WHERE request_id=?`, request_id.ID).
		Scan(&reqIDStr, &moduliJSON, &remaindersJSON, &xStr)

	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, map[string]any{
			"msg":   "未查询到相关结果",
			"found": false,
		})
		return
	}

	//反序列化JSON数组
	var crtParameterString CRTOperationsString
	if err := json.Unmarshal([]byte(moduliJSON), &crtParameterString.Moduli); err != nil {
		http.Error(w, "unmarshal error", http.StatusInternalServerError)
		return
	}
	if err := json.Unmarshal([]byte(remaindersJSON), &crtParameterString.Remainders); err != nil {
		http.Error(w, "unmarshal error", http.StatusInternalServerError)
		return
	}
	crtParameterString.X = xStr
	writeJSON(w, http.StatusOK, crtParameterString)
}

/************* DB *************/
func openDB() (*sql.DB, error) {
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
	return sql.Open("mysql", cfg.FormatDSN())
}
