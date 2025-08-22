package helloworld

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// /api/cert/list
func CertListHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "only support GET", http.StatusMethodNotAllowed)
		return
	}

	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// 一次 JOIN 取回 caPublicKey；timestamp 是列名，建议反引号
	const q = `SELECT l.cert_id, l.serialHex, l.ca_id, IFNULL(c.caPublicKey, '') AS caPublicKey, l.cert_path,
		  l.timestamp FROM dpki.CertList AS l LEFT JOIN dpki.CAList  AS c ON c.ca_id = l.ca_id ORDER BY l.timestamp DESC`

	type row struct {
		CertID      int64     `json:"cert_id"`
		SerialHex   string    `json:"serial_hex"`
		CaID        int64     `json:"ca_id"`
		CAPublicKey string    `json:"caPublicKey"`
		CertPath    string    `json:"cert_path"`
		Ts          time.Time `json:"timestamp"`
	}

	rows, err := db.Query(q)
	if err != nil {
		http.Error(w, "query error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0, 64)
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.CertID, &r.SerialHex, &r.CaID, &r.CAPublicKey, &r.CertPath, &r.Ts); err != nil {
			http.Error(w, "scan error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, map[string]any{
			"cert_id":     r.CertID,
			"serialHex":   r.SerialHex,
			"ca_id":       r.CaID,
			"caPublicKey": r.CAPublicKey,
			"cert_path":   r.CertPath,
			"timestamp":   r.Ts.Format("2006-01-02 15:04:05"),
		})
	}
	fmt.Println(items)
	if err := rows.Err(); err != nil {
		http.Error(w, "rows error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// 下载 Subject 证书
func DownloadSubjectCertHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1) 仅允许 GET
		if r.Method != http.MethodGet {
			http.Error(w, "only support GET method", http.StatusMethodNotAllowed)
			return
		}

		// 2) 解析 id
		idStr := r.URL.Query().Get("cert_id")
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

		var certPath string
		err = db.QueryRowContext(ctx,
			`SELECT cert_path FROM dpki.CertList WHERE cert_id = ?`,
			id,
		).Scan(&certPath)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, err)
			return
		}

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
