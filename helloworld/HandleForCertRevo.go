package helloworld

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

// /api/revoke/list
func RevocationListHandler(w http.ResponseWriter, r *http.Request) {
	db, err := openDB()
	if err != nil {
		http.Error(w, "open db error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	const q = `SELECT r.ID, r.PublicKey, r.ca_id, IFNULL(c.caPublicKey, '') AS caPublicKey, r.Status, r.Description, r.SubjectInfo 
		FROM dpki.requestCertList AS r LEFT JOIN dpki.CAList AS c ON c.ca_id = r.ca_id WHERE r.Status = 'issued' ORDER BY r.ID DESC;`
	//
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

// /api/revoke/cert
func RevocationCertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "only support POST method", http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		ID         int64  `json:"id"`
		X          string `json:"x"`
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
	fmt.Println("test:", in)
	err = db.QueryRow(`SELECT id, request_id FROM requestCertCRT WHERE request_id=? AND x=?`, in.ID, in.X).
		Scan(&in.ID, &in.Request_ID)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	_, err = db.Exec(`UPDATE requestCertList SET Status='revoked' WHERE ID=?`, in.Request_ID)
	if err != nil {
		http.Error(w, "update error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"msg":"certificate revoked"}`))
}
