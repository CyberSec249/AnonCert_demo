package cer_subject_tools

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
)

// 模数请求响应
type ModulusResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Modulus *big.Int `json:"modulus"`
	Error   error    `json:"error,omitempty"`
}

// 模数请求
type ModulusRequest struct {
	SubjectID string `json:"subject_id"`
}

// 余数与主体信息的异或结果请求
type XORRequest struct {
	SubjectID  string   `json:"subject_id"`
	XORResult  []byte   `json:"xor_result"`
	Remainders [][]byte `json:"remainders"`
}

// CRTOperations 包含中国剩余定理操作所需的方法
type CRTOperations struct {
	Subject    *Subject
	Moduli     []*big.Int // 从不同CA获得的模数
	Remainders []*big.Int // 随机生成的余数
	X          *big.Int   // 通过中国剩余定理计算的结果
}

// NewCRTOperations 创建一个新的CRT操作对象
func NewCRTOperations(subject *Subject) *CRTOperations {
	return &CRTOperations{
		Subject:    subject,
		Moduli:     make([]*big.Int, 0),
		Remainders: make([]*big.Int, 0),
	}
}

// RequestModulus 向CA请求模数
func (crt *CRTOperations) RequestModulus(caURL, caName string, subjectID string) (*big.Int, error) {
	modulusRequest := ModulusRequest{
		SubjectID: subjectID,
	}

	jsonData, err := json.Marshal(modulusRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal modulus request: %w", err)
	}

	url := fmt.Sprintf("%s/certificate/modulus/request?caName=%s", caURL, caName)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send modulus request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to request modulus: %s", string(body))
	}

	var modulusResponse ModulusResponse
	if err = json.Unmarshal(body, &modulusResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	if !modulusResponse.Success {
		return nil, fmt.Errorf("modulus request failed: %s", modulusResponse.Message)
	}

	return modulusResponse.Modulus, nil
}

// RequestAllModuli 向多个CA请求模数
func (crt *CRTOperations) RequestAllModuli(caURLs []string, caNames []string, subjectID string) error {
	if len(caURLs) != len(caNames) {
		return fmt.Errorf("CA URLs 和 CA Names 数量不匹配")
	}

	crt.Moduli = make([]*big.Int, len(caURLs))

	for i := range caURLs {
		modulus, err := crt.RequestModulus(caURLs[i], caNames[i], subjectID)
		if err != nil {
			return err
		}
		crt.Moduli[i] = modulus
	}

	return nil
}

// GenerateRandomRemainders 为每个模数生成随机余数
func (crt *CRTOperations) GenerateRandomRemainders() error {
	if len(crt.Moduli) == 0 {
		return fmt.Errorf("模数列表为空，请先获取模数")
	}

	crt.Remainders = make([]*big.Int, len(crt.Moduli))

	for i, modulus := range crt.Moduli {
		// 生成小于模数的随机整数作为余数
		remainder, err := rand.Int(rand.Reader, modulus)
		if err != nil {
			return fmt.Errorf("生成随机余数时出错: %w", err)
		}
		crt.Remainders[i] = remainder
	}

	return nil
}

// SolveChineseRemainderTheorem 实现中国剩余定理求解x
func (crt *CRTOperations) SolveChineseRemainderTheorem() error {
	if len(crt.Moduli) == 0 || len(crt.Remainders) == 0 {
		return fmt.Errorf("模数或余数列表为空")
	}

	if len(crt.Moduli) != len(crt.Remainders) {
		return fmt.Errorf("模数和余数数量不匹配")
	}

	// 计算所有模数的乘积 M
	M := big.NewInt(1)
	for _, modulus := range crt.Moduli {
		M.Mul(M, modulus)
	}

	// 计算 x = Σ(r_i * M_i * M_i^(-1) mod n_i) mod M
	x := big.NewInt(0)

	for i := 0; i < len(crt.Moduli); i++ {
		// M_i = M / n_i
		Mi := new(big.Int).Div(M, crt.Moduli[i])

		// M_i^(-1) mod n_i
		MiInv := new(big.Int).ModInverse(Mi, crt.Moduli[i])

		if MiInv == nil {
			return fmt.Errorf("模数 %v 不是互素的", crt.Moduli[i])
		}

		// r_i * M_i * M_i^(-1)
		term := new(big.Int).Mul(crt.Remainders[i], Mi)
		term.Mul(term, MiInv)

		// 累加
		x.Add(x, term)
	}

	// 对 M 取模
	crt.X = new(big.Int).Mod(x, M)

	return nil
}

// XORWithSubjectInfo 将计算得到的x与主体信息进行异或运算
func (crt *CRTOperations) XORWithSubjectInfo(subjectInfoBytes []byte) ([]byte, error) {

	if len(crt.Remainders) == 0 {
		return nil, fmt.Errorf("余数列表为空，请先生成随机余数")
	}

	xorResult := make([]byte, len(subjectInfoBytes))

	copy(xorResult, subjectInfoBytes)

	for _, remainder := range crt.Remainders {
		remainderBytes := remainder.Bytes()

		for i := 0; i < len(xorResult); i++ {
			xorResult[i] = xorResult[i] ^ remainderBytes[i%len(remainderBytes)]
		}
	}

	return xorResult, nil
}

// SendXORResultToCA 将异或结果和余数发送给CA
func (crt *CRTOperations) SendXORResultToCA(caURL string, caName string, subjectID string, xorResult []byte) error {
	remainderBytes := make([][]byte, len(crt.Remainders))
	for i, remainder := range crt.Remainders {
		remainderBytes[i] = remainder.Bytes()
	}

	xorRequest := XORRequest{
		SubjectID:  subjectID,
		XORResult:  xorResult,
		Remainders: remainderBytes,
	}

	jsonData, err := json.Marshal(xorRequest)
	if err != nil {
		return fmt.Errorf("failed to marshal XOR request: %w", err)
	}

	url := fmt.Sprintf("%s/certificate/xorResult?caName=%s", caURL, caName)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send XOR result: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send XOR result: %s", string(body))
	}

	return nil
}
