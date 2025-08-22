package helloworld

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
)

type CRTOperations struct {
	ID         int64      `json:"id"`
	Moduli     []*big.Int `json:"moduli"`     // 从不同CA获得的模数
	Remainders []*big.Int `json:"remainders"` // 随机生成的余数
	X          *big.Int   `json:"x"`          // 通过中国剩余定理计算的结果
}

type CRTOperationsString struct {
	ID         int64    `json:"id"`
	Moduli     []string `json:"moduli"`     // 从不同CA获得的模数
	Remainders []string `json:"remainders"` // 随机生成的余数
	X          string   `json:"x"`          // 通过中国剩余定理计算的结果
}

/************* PrimePool *************/
type PrimePool struct {
	Primes []*big.Int
	seen   map[string]struct{}
}

func NewPrimePool() *PrimePool {
	return &PrimePool{
		Primes: make([]*big.Int, 0),
		seen:   make(map[string]struct{}),
	}
}

// 生成指定数量的质数并加入池中（去重）
func (pool *PrimePool) GeneratePrimes(count, bits int) error {
	for i := 0; i < count; {
		prime, err := rand.Prime(rand.Reader, bits)
		if err != nil {
			return fmt.Errorf("生成质数时出错: %w", err)
		}
		key := prime.Text(16)
		if _, ok := pool.seen[key]; ok {
			continue
		}
		pool.seen[key] = struct{}{}
		pool.Primes = append(pool.Primes, prime)
		i++
	}
	return nil
}

// 从池中随机选择一个质数
func (pool *PrimePool) GetRandomPrime() (*big.Int, error) {
	if len(pool.Primes) == 0 {
		return nil, fmt.Errorf("质数池为空")
	}
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool.Primes))))
	if err != nil {
		return nil, fmt.Errorf("随机选择质数时出错: %w", err)
	}
	return pool.Primes[index.Int64()], nil
}

// 从池中随机选择 k 个的质数
func (pool *PrimePool) RandomModuli(k int) ([]*big.Int, error) {
	n := len(pool.Primes)
	if k <= 0 {
		return nil, fmt.Errorf("k 必须 > 0")
	}
	if n < k {
		return nil, fmt.Errorf("质数池不足: have=%d need=%d", n, k)
	}
	out := make([]*big.Int, 0, k)
	used := make(map[int64]struct{})
	for len(out) < k {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
		if err != nil {
			return nil, err
		}
		i := idx.Int64()
		if _, ok := used[i]; ok {
			continue
		}
		used[i] = struct{}{}
		out = append(out, new(big.Int).Set(pool.Primes[i])) // 拷贝一份
	}
	return out, nil
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
func (crt *CRTOperations) XORWithSubjectInfo(subjectInfoBytes []byte) []byte {

	xorResult := make([]byte, len(subjectInfoBytes))

	copy(xorResult, subjectInfoBytes)

	for _, remainder := range crt.Remainders {
		remainderBytes := remainder.Bytes()

		for i := 0; i < len(xorResult); i++ {
			xorResult[i] = xorResult[i] ^ remainderBytes[i%len(remainderBytes)]
		}
	}

	return xorResult
}

func ConvertCRT(raw CRTOperationsString) (*CRTOperations, error) {
	result := &CRTOperations{
		ID:         raw.ID,
		Moduli:     make([]*big.Int, len(raw.Moduli)),
		Remainders: make([]*big.Int, len(raw.Remainders)),
	}

	// 转换 moduli
	for i, s := range raw.Moduli {
		n, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("invalid moduli[%d]: %q", i, s)
		}
		result.Moduli[i] = n
	}

	// 转换 remainders
	for i, s := range raw.Remainders {
		r, ok := new(big.Int).SetString(s, 10)
		if !ok {
			return nil, fmt.Errorf("invalid remainders[%d]: %q", i, s)
		}
		result.Remainders[i] = r
	}

	// 转换 X
	if raw.X != "" {
		x, ok := new(big.Int).SetString(raw.X, 10)
		if !ok {
			return nil, fmt.Errorf("invalid x: %q", raw.X)
		}
		result.X = x
	} else {
		result.X = big.NewInt(0) // 或者保持 nil
	}

	return result, nil
}

func ConvertToStringVersion(src CRTOperations) CRTOperationsString {
	dst := CRTOperationsString{
		ID:         src.ID,
		Moduli:     make([]string, len(src.Moduli)),
		Remainders: make([]string, len(src.Remainders)),
	}

	for i, m := range src.Moduli {
		if m != nil {
			dst.Moduli[i] = m.String() // 十进制字符串
		}
	}

	for i, r := range src.Remainders {
		if r != nil {
			dst.Remainders[i] = r.String() // 十进制字符串
		}
	}

	if src.X != nil {
		dst.X = src.X.String()
	}

	return dst
}

func parseBigInt(s string) (*big.Int, error) {
	s = strings.TrimSpace(s)
	// 支持 0x 开头
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		z, ok := new(big.Int).SetString(s[2:], 16)
		if !ok {
			return nil, fmt.Errorf("invalid hex big int: %q", s)
		}
		return z, nil
	}
	z, ok := new(big.Int).SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("invalid decimal big int: %q", s)
	}
	return z, nil
}

func parseSlice(ss []string) ([]*big.Int, error) {
	out := make([]*big.Int, 0, len(ss))
	for i, s := range ss {
		z, err := parseBigInt(s)
		if err != nil {
			return nil, fmt.Errorf("moduli/remainders[%d]: %w", i, err)
		}
		out = append(out, z)
	}
	return out, nil
}

// ValidateCRT 验证 X 是否同时满足 X ≡ rᵢ (mod nᵢ)
// 额外做了稳健性校验：长度一致、nᵢ>1、0≤rᵢ<nᵢ。
// 若你需要，还可打开“互素性检查”，确保解的唯一性（可选）。
func ValidateCRT(ops CRTOperations) (ok bool, detail string) {
	// 1) 基础校验
	if len(ops.Moduli) == 0 || len(ops.Remainders) == 0 {
		return false, "moduli/remainders 为空"
	}
	if len(ops.Moduli) != len(ops.Remainders) {
		return false, "moduli 与 remainders 长度不一致"
	}
	if ops.X == nil {
		return false, "X 为空"
	}

	// 2) 可选：检查模数两两互素（唯一解条件；不互素也可存在解，但不保证唯一）
	// 如果你需要强制唯一性，取消注释这段

	if okPairwise, i, j := pairwiseCoprime(ops.Moduli); !okPairwise {
		return false, fmt.Sprintf("模数不两两互素：gcd(n%d, n%d) ≠ 1", i+1, j+1)
	}

	// 3) 逐一验证 X ≡ rᵢ (mod nᵢ)
	tmpX := new(big.Int)
	tmpR := new(big.Int)
	for i := range ops.Moduli {
		n := ops.Moduli[i]
		r := ops.Remainders[i]

		// nᵢ 合法性
		if n == nil || n.Sign() <= 0 {
			return false, fmt.Sprintf("n%d 非法（应为正整数）", i+1)
		}
		if n.Cmp(big.NewInt(1)) <= 0 {
			return false, fmt.Sprintf("n%d 必须 > 1", i+1)
		}
		// rᵢ 合法性：0 ≤ rᵢ < nᵢ
		if r == nil || r.Sign() < 0 {
			return false, fmt.Sprintf("r%d 非法（应为非负整数）", i+1)
		}
		if r.Cmp(n) >= 0 {
			return false, fmt.Sprintf("r%d 必须满足 0 ≤ r%d < n%d", i+1, i+1, i+1)
		}

		// 计算 X mod nᵢ 与 rᵢ（必要时把 rᵢ 也约简到 mod nᵢ）
		xMod := tmpX.Mod(ops.X, n)
		rMod := tmpR.Mod(r, n)
		if xMod.Cmp(rMod) != 0 {
			return false, fmt.Sprintf("在 i=%d 处不成立：X mod n%d = %s，r%d = %s",
				i+1, i+1, xMod.String(), i+1, rMod.String())
		}
	}

	return true, "所有同余均成立"
}

// 可选：检查模数两两互素（用于保证唯一解）
// 返回是否互素，以及第一个不互素的下标对 (i, j)
func pairwiseCoprime(mods []*big.Int) (bool, int, int) {
	g := new(big.Int)
	for i := 0; i < len(mods); i++ {
		for j := i + 1; j < len(mods); j++ {
			if g.GCD(nil, nil, mods[i], mods[j]).Cmp(big.NewInt(1)) != 0 {
				return false, i, j
			}
		}
	}
	return true, -1, -1
}
