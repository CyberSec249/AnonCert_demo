package cer_ca_tools

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// 互素大质数池
type PrimePool struct {
	Primes []*big.Int
}

// 初始化质数池
func NewPrimePool() *PrimePool {
	return &PrimePool{
		Primes: make([]*big.Int, 0),
	}
}

// 生成指定数量的质数并加入池中
func (pool *PrimePool) GeneratePrimes(count, bits int) error {
	for i := 0; i < count; {
		prime, err := GeneratePrime(bits)
		if err != nil {
			return fmt.Errorf("生成质数时出错: %w", err)
		}

		// 只有当新生成的质数与池中所有质数互质时，才添加到池中
		if pool.AddPrime(prime) {
			i++ // 只有成功添加时才递增计数器
		}
	}
	return nil
}

// 生成指定位数的质数
func GeneratePrime(bits int) (*big.Int, error) {
	return rand.Prime(rand.Reader, bits)
}

// 检查是否与现有质数互质
func (pool *PrimePool) IsCoprime(prime *big.Int) bool {
	for _, existingPrime := range pool.Primes {
		// 计算最大公约数
		gcd := new(big.Int).GCD(nil, nil, prime, existingPrime)
		// 如果最大公约数不为1，表示不互质
		if gcd.Cmp(big.NewInt(1)) != 0 {
			return false
		}
	}
	return true
}

// 向池中添加质数，确保互质
func (pool *PrimePool) AddPrime(prime *big.Int) bool {
	if pool.IsCoprime(prime) {
		pool.Primes = append(pool.Primes, prime)
		return true
	}
	return false
}

// 从池中随机选择一个质数
func (pool *PrimePool) GetRandomPrime() (*big.Int, error) {
	if len(pool.Primes) == 0 {
		return nil, fmt.Errorf("质数池为空")
	}

	// 随机选择一个索引
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool.Primes))))
	if err != nil {
		return nil, fmt.Errorf("随机选择质数时出错: %w", err)
	}

	return pool.Primes[index.Int64()], nil
}
