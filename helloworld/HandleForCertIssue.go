package helloworld

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type IssueCertListInfo struct {
	PublicKey      string `json:"publicKey"`
	Request_ID     int64  `json:"request_id"`
	Subject_Status string `json:"subject_status"`
	Crt_Status     string `json:"crt_status"`
	IFReject       string `json:"if_reject"`
}

// /api/issue/list
func IssueListHandler(w http.ResponseWriter, r *http.Request) {
	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT request_id, subject_status, crt_status, if_reject FROM issueCertList`)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []IssueCertListInfo
	for rows.Next() {
		var item IssueCertListInfo
		if err := rows.Scan(&item.Request_ID, &item.Subject_Status, &item.Crt_Status, &item.IFReject); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		err = db.QueryRow(`SELECT PublicKey
                       FROM requestCertList WHERE id=?`, item.Request_ID).Scan(&item.PublicKey)
		list = append(list, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": list})
}

func IssueSubjectInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		ID string `json:"id"`
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

	var out struct {
		SubjectInfo string `json:"subjectInfo"`
	}
	err = db.QueryRow(`SELECT SubjectInfo FROM requestCertList WHERE id=?`, in.ID).Scan(&out.SubjectInfo)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// /api/issue/crt/detail
func IssueCRTQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		Request_ID string `json:"request_id"`
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

	var (
		reqIDStr       string
		moduliJSON     string
		remaindersJSON string
		xStr           string
	)
	err = db.QueryRow(`SELECT request_id, moduli, remainders, x
                       FROM requestCertCRT WHERE request_id=?`, in.Request_ID).
		Scan(&reqIDStr, &moduliJSON, &remaindersJSON, &xStr)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusOK, map[string]any{
			"msg":   "证书主体尚未提交CRT参数",
			"found": false,
		})
		return
	}

	//反序列化JSON数组
	var crtParameterString CRTOperationsString
	crtParameterString.ID, _ = strconv.ParseInt(reqIDStr, 10, 64)
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

// /api/issue/cert/check
func IssueSubjectCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}

	var issueCertStatus IssueCertListInfo
	if err := json.NewDecoder(r.Body).Decode(&issueCertStatus); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if issueCertStatus.Subject_Status != "" && issueCertStatus.Crt_Status == "" {
		_, err = db.Exec(
			`UPDATE issueCertList SET subject_status = ? WHERE request_id =?`, issueCertStatus.Subject_Status, issueCertStatus.Request_ID,
		)

		_, err = db.Exec(
			`UPDATE requestCertList SET Status = 'waitCRT' WHERE id =?`, issueCertStatus.Request_ID,
		)
	} else if issueCertStatus.Subject_Status == "" && issueCertStatus.Crt_Status != "" {
		_, err = db.Exec(
			`UPDATE issueCertList SET crt_status = ? WHERE request_id =?`, issueCertStatus.Crt_Status, issueCertStatus.Request_ID,
		)

		_, err = db.Exec(
			`UPDATE requestCertList SET Status = 'acceptCRT' WHERE id =?`, issueCertStatus.Request_ID,
		)
	} else {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(w, "update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// /api/issue/cert/reject
func IssueSubjectRejectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}

	var in struct {
		Request_ID int64 `json:"request_id"`
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

	fmt.Println(in.Request_ID)
	_, err = db.Exec(`UPDATE issueCertList SET if_reject = 1 WHERE request_id =?`, in.Request_ID)
	_, err = db.Exec(`UPDATE requestCertList SET Status = 'rejected' WHERE id =?`, in.Request_ID)
	if err != nil {
		http.Error(w, "update error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// /api/issue/crt/verify
func IssueCRTVerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}

	var crtParameterString CRTOperationsString
	if err := json.NewDecoder(r.Body).Decode(&crtParameterString); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var err error
	var crtParameter CRTOperations
	crtParameter.ID = crtParameterString.ID
	if crtParameter.Moduli, err = parseSlice(crtParameterString.Moduli); err != nil {
		http.Error(w, "invalid moduli: "+err.Error(), http.StatusBadRequest)
		return
	}
	if crtParameter.Remainders, err = parseSlice(crtParameterString.Remainders); err != nil {
		http.Error(w, "invalid remainders: "+err.Error(), http.StatusBadRequest)
		return
	}
	if crtParameter.X, err = parseBigInt(crtParameterString.X); err != nil {
		http.Error(w, "invalid x: "+err.Error(), http.StatusBadRequest)
		return
	}

	ok, detail := ValidateCRT(crtParameter)
	fmt.Println(ok, detail)

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	w.WriteHeader(http.StatusOK)
}

// /api/issue/cert/issuance
func IssueCertIssuanceHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, "only support POST method")
		return
	}
	defer r.Body.Close()

	var in struct {
		Request_ID int64 `json:"request_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeJSON(w, http.StatusBadRequest, err)
		return
	}

	db, err := openDB()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err)
		return
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()

	/***** 1) 取该请求的 CRT 参数（字符串 JSON → 结构体） *****/
	var reqIDStr, moduliJSON, remaindersJSON, xStr string
	err = db.QueryRowContext(ctx,
		`SELECT request_id, moduli, remainders, x FROM dpki.requestCertCRT WHERE request_id=?`,
		in.Request_ID,
	).Scan(&reqIDStr, &moduliJSON, &remaindersJSON, &xStr)
	if err != nil {
		writeJSON(w, http.StatusNotFound, "CRT parameters not found")
		return
	}
	crtStr := CRTOperationsString{ID: in.Request_ID, X: xStr}
	if err := json.Unmarshal([]byte(moduliJSON), &crtStr.Moduli); err != nil {
		writeJSON(w, http.StatusInternalServerError, "unmarshal moduli failed")
		return
	}
	if err := json.Unmarshal([]byte(remaindersJSON), &crtStr.Remainders); err != nil {
		writeJSON(w, http.StatusInternalServerError, err)
		return
	}
	crtParam, err := ConvertCRT(crtStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, "convert CRT failed")
		return
	}

	/***** 2) 构造匿名主题（使用 XORWithSubjectInfo） *****/
	origin := pkix.Name{
		Country:            []string{"CN"},
		Province:           []string{"Beijing"},
		Locality:           []string{"Beijing"},
		Organization:       []string{"Test Client"},
		OrganizationalUnit: []string{"IT"},
		CommonName:         "Test Subject",
	}
	anon, err := buildAnonymousSubject(origin, crtParam.XORWithSubjectInfo)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "build anonymous subject failed")
		return
	}

	/***** 3) 查申请者公钥 + 负责该请求的 CA（证书与密钥路径） *****/
	var applicantPub string
	if err := db.QueryRowContext(ctx, `
		SELECT PublicKey FROM dpki.requestCertList WHERE ID = ?`, in.Request_ID).Scan(&applicantPub); err != nil {
		writeJSON(w, http.StatusNotFound, "public key not found")
		return
	}

	ca, err := findCAForRequest(ctx, db, in.Request_ID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, "CA not found for this request")
		return
	}
	fmt.Println(ca)
	caCert, caKey, err := loadCACredential(ca.CertPath, ca.KeyPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "load CA credential failed")
		return
	}

	//pub, err := parseApplicantPubKey(applicantPub)
	//if err != nil {
	//	writeJSON(w, http.StatusBadRequest, "invalid applicant public key")
	//	return
	//}

	if err := os.MkdirAll("./helloworld/subCert", 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, "创建证书目录失败")
		return
	}
	if err := os.MkdirAll("./helloworld/subKey", 0o700); err != nil {
		writeJSON(w, http.StatusInternalServerError, "创建私钥目录失败")
		return
	}

	subPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "生成CA椭圆曲线私钥失败")
		return
	}

	/***** 4) 生成并签发证书 *****/
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	leafSerial, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "gen serial failed")
		return
	}

	leaf := x509.Certificate{
		SerialNumber:          leafSerial,
		Subject:               anon, // 匿名化主题
		NotBefore:             time.Now().Add(-5 * time.Minute),
		NotAfter:              time.Now().AddDate(1, 0, 0), // 有效期 1 年
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	leafDER, err := x509.CreateCertificate(rand.Reader, &leaf, caCert, &subPrivKey.PublicKey, caKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "create certificate failed")
		return
	}
	leafPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leafDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(subPrivKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "序列化CA私钥失败")
		return
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	/***** 5) 落盘 + 入库（CertList） + 更新申请状态 *****/
	if err := os.MkdirAll("./helloworld/issuedCert", 0o755); err != nil {
		writeJSON(w, http.StatusInternalServerError, "mkdir issuedCert failed")
		return
	}
	serialHex := strings.ToUpper(leafSerial.Text(16))
	outPath := filepath.Join("./helloworld/subCert", fmt.Sprintf("req-%d-%s.crt", in.Request_ID, serialHex))
	keyPath := filepath.Join("./helloworld/subKey", fmt.Sprintf("sub-%s.key", serialHex))

	if err := os.WriteFile(outPath, leafPEM, 0o644); err != nil {
		writeJSON(w, http.StatusInternalServerError, "写入主体证书失败")
		return
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		writeJSON(w, http.StatusInternalServerError, "写入主体私钥失败")
		return
	}

	// —— 计算公钥字符串（未压缩点 HEX：04 || X || Y）
	//pubBytes := elliptic.Marshal(elliptic.P384(), subPrivKey.PublicKey.X, subPrivKey.PublicKey.Y)
	//pubHex := strings.ToUpper(hex.EncodeToString(pubBytes)) // 例：04A3...

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "begin tx failed")
		return
	}
	defer tx.Rollback()

	// 1) 写入 CertList（根据你的表结构微调列名）
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO dpki.CertList (ca_id,serialHex, cert_path) VALUES (?, ?, ?)`, ca.CA_ID, serialHex, outPath); err != nil {
		writeJSON(w, http.StatusInternalServerError, "insert CertList failed")
		return
	}
	// 2) 更新申请单状态
	if _, err := tx.ExecContext(ctx,
		`UPDATE dpki.requestCertList SET status='issued' WHERE ID=?`,
		in.Request_ID,
	); err != nil {
		writeJSON(w, http.StatusInternalServerError, "update request status failed")
		return
	}
	if err := tx.Commit(); err != nil {
		writeJSON(w, http.StatusInternalServerError, "commit tx failed")
		return
	}

	/***** 6) 返回 JSON 结果 *****/
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"msg":          "issued",
		"request_id":   in.Request_ID,
		"serial_hex":   serialHex,
		"cert_path":    outPath,
		"ca_id":        ca.CA_ID,
		"ca_cert_path": ca.CertPath,
	})
}
