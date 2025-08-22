package helloworld

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CAInfo struct {
	CA_ID       string `json:"ca_id"`
	CAPublicKey string `json:"caPublicKey"`
	CAStatus    string `json:"caStatus"`
	CertPath    string `json:"cert_path"`
	KeyPath     string `json:"key_path"`
}

// /api/ca/list
func CAListHandler(w http.ResponseWriter, r *http.Request) {
	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(`SELECT ca_id, caPublicKey, caStatus, cert_path FROM CAList`)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var caList []CAInfo
	for rows.Next() {
		var caItem CAInfo
		if err := rows.Scan(&caItem.CA_ID, &caItem.CAPublicKey, &caItem.CAStatus, &caItem.CertPath); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		caList = append(caList, caItem)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": caList})
}

// /api/ca/register
func GenerateCAHandle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	fail := func(code int, msg string, err error) {
		if err != nil {
			log.Printf("[GenerateCA] %s: %v", msg, err)
		}
		http.Error(w, msg, code)
	}

	if err := os.MkdirAll("./helloworld/caCert", 0o755); err != nil {
		fail(http.StatusInternalServerError, "创建证书目录失败", err)
		return
	}
	if err := os.MkdirAll("./helloworld/caKey", 0o700); err != nil {
		fail(http.StatusInternalServerError, "创建私钥目录失败", err)
		return
	}

	caPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		fail(http.StatusInternalServerError, "生成CA椭圆曲线私钥失败", err)
		return
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 160)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fail(http.StatusInternalServerError, "生成序列号失败", err)
		return
	}

	tmpl := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Province:           []string{"Beijing"},
			Locality:           []string{"Beijing"},
			Organization:       []string{"Test CA"},
			OrganizationalUnit: []string{"IT"},
			CommonName:         "Test_CA_one",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            2,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fail(http.StatusInternalServerError, "生成CA证书失败", err)
		return
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(caPrivKey)
	if err != nil {
		fail(http.StatusInternalServerError, "序列化CA私钥失败", err)
		return
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	// 唯一文件名：使用序列号十六进制
	serialHex := strings.ToUpper(serialNumber.Text(16))
	certPath := filepath.Join("./helloworld/caCert", fmt.Sprintf("%s-%s.crt", tmpl.Subject.CommonName, serialHex))
	keyPath := filepath.Join("./helloworld/caKey", fmt.Sprintf("ca-%s.key", serialHex))

	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		fail(http.StatusInternalServerError, "写入CA证书失败", err)
		return
	}
	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		fail(http.StatusInternalServerError, "写入CA私钥失败", err)
		return
	}

	// —— 计算公钥字符串（未压缩点 HEX：04 || X || Y）
	pubBytes := elliptic.Marshal(elliptic.P384(), caPrivKey.PublicKey.X, caPrivKey.PublicKey.Y)
	pubHex := strings.ToUpper(hex.EncodeToString(pubBytes)) // 例：04A3...

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// —— 写入数据库（状态默认 active）
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	const insertSQL = `INSERT INTO dpki.CAList (caPublicKey, caStatus, cert_path, key_path) VALUES (?, ?, ?, ?)`
	res, err := db.ExecContext(ctx, insertSQL, pubHex, "active", certPath, keyPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, "写入数据库失败")
		return
	}

	newID, _ := res.LastInsertId()

	// 成功返回 JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated) // 201
	_ = json.NewEncoder(w).Encode(map[string]any{
		"msg":        "ok",
		"id":         newID,
		"serial":     serialNumber.String(),
		"serial_hex": serialHex,
		"public_key": pubHex,
		"status":     "active",
		"cert_path":  certPath,
		// 如需前端下载私钥，绝不要直返私钥；仅返回 keyPath/受控下载接口
		// "key_path": keyPath,
	})
}

// 下载 CA 证书：根据 ca_id 返回证书文件（PEM）
func DownloadCACertHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 仅允许 GET
		if r.Method != http.MethodGet {
			http.Error(w, "only support GET method", http.StatusMethodNotAllowed)
			return
		}

		// 2) 解析 id
		idStr := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}

		db, err := openDB()
		if err != nil {
			http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// 3) 查库拿证书路径（并可校验状态）
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var certPath, status string
		err = db.QueryRowContext(ctx,
			`SELECT cert_path, caStatus FROM dpki.CAList WHERE ca_id = ?`,
			id,
		).Scan(&certPath, &status)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}

		// 可选：撤销状态不允许下载（按需放开）
		// if status == "revoked" {
		// 	writeJSONError(w, http.StatusForbidden, "certificate revoked", nil)
		// 	return
		// }

		// 4) 打开文件并发送
		f, err := os.Open(certPath)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil || fi.IsDir() {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}

		filename := filepath.Base(certPath) // 例如 Test_CA_one-<serial>.crt
		// 证书为 PEM：可用 application/x-pem-file；若 DER 可用 application/pkix-cert
		w.Header().Set("Content-Type", "application/x-pem-file")
		// 触发下载保存对话框
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
		w.Header().Set("Last-Modified", fi.ModTime().UTC().Format(http.TimeFormat))
		// 按需：w.Header().Set("Cache-Control", "no-store") // 避免缓存

		// 高效/支持范围请求的输出
		http.ServeContent(w, r, filename, fi.ModTime(), f)
	}
}
