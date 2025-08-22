package helloworld

import (
	"io/ioutil"
	"log"
)

func mustRead(path string) string {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

//func main() {
//	// 1) 连接 FISCO-BCOS
//	cfgFile := "./config.toml" // 你的配置文件
//	c, err := client.DialWithConfig(cfgFile)
//	if err != nil {
//		log.Fatalf("dial fisco: %v", err)
//	}
//
//	// 2) 读取 ABI / BIN
//	abiJSON := mustRead("./build/CertOper.abi")
//	binHex := mustRead("./build/CertOper.bin") // 无构造参数，直接用
//
//	// 3) 部署
//	// 不同 SDK 版本：函数名可能是 DeployContract / DeployContractWithSign / DeployContractWithReceipt 等
//	// 这里给出常见签名：address, txHash, err := c.DeployContract(abi, bin, params...)
//	// 如果你使用的是带 receipt 的版本，会返回 receipt，里边能拿到合约地址
//	contractAddr, txHash, err := c.DeployContract(abiJSON, binHex) // 无构造参数
//	if err != nil {
//		log.Fatalf("deploy: %v", err)
//	}
//	fmt.Println("TX:", txHash)
//	fmt.Println("Contract Address:", contractAddr)
//
//	// 4) 等待上链（可选）
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	receipt, err := c.GetTransactionReceipt(ctx, txHash)
//	if err != nil {
//		log.Printf("receipt err: %v", err)
//	} else {
//		fmt.Println("Receipt status:", receipt.Status)
//	}
//}
