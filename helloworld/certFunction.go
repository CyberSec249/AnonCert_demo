package helloworld

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"crypto/x509/pkix"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

type xorFunc func([]byte) []byte

// 将任意可 JSON 的值：v -> JSON -> XOR -> Base64 字符串
func maskJSONToB64(v any, xor xorFunc) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal %T failed: %w", v, err)
	}
	c := xor(b)
	return base64.StdEncoding.EncodeToString(c), nil
}

// 将 string：s -> UTF-8 bytes -> XOR -> Base64 字符串
func maskStringToB64(s string, xor xorFunc) string {
	c := xor([]byte(s))
	return base64.StdEncoding.EncodeToString(c)
}

// 基于上面的工具，生成匿名化的 pkix.Name
func buildAnonymousSubject(subject pkix.Name, xor xorFunc) (pkix.Name, error) {
	countryEnc, err := maskJSONToB64(subject.Country, xor)
	if err != nil {
		return pkix.Name{}, err
	}
	provinceEnc, err := maskJSONToB64(subject.Province, xor)
	if err != nil {
		return pkix.Name{}, err
	}
	localityEnc, err := maskJSONToB64(subject.Locality, xor)
	if err != nil {
		return pkix.Name{}, err
	}
	orgEnc, err := maskJSONToB64(subject.Organization, xor)
	if err != nil {
		return pkix.Name{}, err
	}
	ouEnc, err := maskJSONToB64(subject.OrganizationalUnit, xor)
	if err != nil {
		return pkix.Name{}, err
	}

	anon := pkix.Name{
		CommonName:         maskStringToB64(subject.CommonName, xor),
		Country:            []string{countryEnc},
		Province:           []string{provinceEnc},
		Locality:           []string{localityEnc},
		Organization:       []string{orgEnc},
		OrganizationalUnit: []string{ouEnc},
	}
	return anon, nil
}

/********** 辅助函数：解析申请者公钥（PEM 或 HEX 未压缩点） **********/
func parseApplicantPubKey(s string) (any, error) {
	ss := strings.TrimSpace(s)
	if ss == "" {
		return nil, errors.New("empty public key")
	}
	// PEM?
	if strings.HasPrefix(ss, "-----BEGIN") {
		block, _ := pem.Decode([]byte(ss))
		if block == nil {
			return nil, errors.New("invalid PEM")
		}
		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("parse PKIX public key failed: %w", err)
		}
		return pub, nil
	}
	// HEX 未压缩点（04 || X || Y）
	raw, err := hex.DecodeString(ss)
	if err != nil {
		return nil, fmt.Errorf("decode hex failed: %w", err)
	}
	if len(raw) < 2 || raw[0] != 0x04 {
		return nil, errors.New("expect uncompressed EC point (starts with 0x04)")
	}
	// 根据长度猜曲线（除去前导 0x04）
	l := len(raw) - 1
	var curve elliptic.Curve
	switch l {
	case 64: // 32+32
		curve = elliptic.P256()
	case 96: // 48+48
		curve = elliptic.P384()
	case 132: // 66+66
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported EC point length: %d", len(raw))
	}
	x, y := elliptic.Unmarshal(curve, raw)
	if x == nil || y == nil {
		return nil, errors.New("elliptic.Unmarshal failed")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

/********** 辅助函数：从 DB 定位该请求使用的 CA（优先按 CA 公钥关联） **********/
type caRow struct {
	CA_ID       int64  `json:"ca_id"`
	CAStatus    string `json:"caStatus"`
	CAPublicKey string `json:"caPublicKey"`
	CertPath    string `json:"certPath"`
	KeyPath     string `json:"keyPath"`
}

func findCAForRequest(ctx context.Context, db *sql.DB, reqID int64) (caRow, error) {
	var row caRow
	fmt.Println(reqID)
	qB := `SELECT c.ca_id, c.caStatus, c.caPublicKey, c.cert_path, c.key_path FROM dpki.requestCertList r 
    JOIN dpki.CAList c ON c.ca_id = r.ca_id WHERE r.id = ? LIMIT 1`
	err := db.QueryRowContext(ctx, qB, reqID).Scan(&row.CA_ID, &row.CAStatus, &row.CAPublicKey, &row.CertPath, &row.KeyPath)

	return row, err
}

/********** 辅助函数：读取并解析 CA 证书与私钥（PKCS#8） **********/
func loadCACredential(certPath, keyPath string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	fmt.Println(certPath, keyPath)
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read ca cert: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, errors.New("invalid ca cert pem")
	}

	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse ca cert: %w", err)
	}

	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read ca key: %w", err)
	}
	kb, _ := pem.Decode(keyPEM)
	if kb == nil {
		return nil, nil, errors.New("invalid ca key pem")
	}
	privAny, err := x509.ParsePKCS8PrivateKey(kb.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse pkcs8 key: %w", err)
	}
	priv, ok := privAny.(*ecdsa.PrivateKey)
	if !ok {
		return nil, nil, errors.New("ca key is not ECDSA")
	}
	return caCert, priv, nil
}
