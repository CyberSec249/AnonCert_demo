package helloworld

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/FISCO-BCOS/go-sdk/client"
	"github.com/FISCO-BCOS/go-sdk/conf"
	"github.com/ethereum/go-ethereum/common"
)

func BlockInfoGet() (blockNum string, txNum string, nodeNum string, contractNum string) {

	ctx := context.Background()

	configs, err := conf.ParseConfigFile("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	config := &configs[0]

	c, err := client.Dial(config)
	if err != nil {
		log.Fatal(err)
	}

	// load the HelloWorld = 0xAAC410d4093Ad00dc6995f853864701b3b71845E
	contractAddress := common.HexToAddress("0xAAC410d4093Ad00dc6995f853864701b3b71845E")
	instance, err := NewHelloWorld(contractAddress, c)
	if err != nil {
		log.Fatal(err)
	}

	helloworldSession := &HelloWorldSession{Contract: instance, CallOpts: *c.GetCallOpts(), TransactOpts: *c.GetTransactOpts()}

	value, err := helloworldSession.Get() // call Get API
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("value :", value)

	//value = "Hello, FISCO BCOS"
	//tx, receipt, err := helloworldSession.Set(value) // call set API
	//if err != nil {
	//	log.Fatal(err)
	//}

	//fmt.Printf("tx sent: %s\n", tx.Hash().Hex())
	//fmt.Printf("transaction hash of receipt: %s\n", receipt.GetTransactionHash())

	blockNumInt, err := c.GetBlockNumber(ctx)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("block number: %d\n", blockNumInt)

	created, err := countNewContracts(ctx, c, blockNumInt, 200) // 200 可按需调整
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("contract number: %d\n", created)

	txs, err := c.GetTotalTransactionCount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	txNumInt, _ := strconv.ParseInt(txs.TxSum[2:], 16, 64)
	fmt.Printf("tx number: %d\n", txNumInt)

	blockNum = strconv.FormatInt(blockNumInt, 10)
	txNum = strconv.FormatInt(txNumInt, 10)
	contractNum = strconv.FormatInt(created, 10)
	peers, err := c.GetPeers(ctx) // 如需节点数，可用 GetPeers(ctx) 再 len()
	if err != nil {
		log.Fatal(err)
	}
	nodeNum = strconv.FormatInt(int64(len(*peers)+1), 10)
	return
}

func countNewContracts(ctx context.Context, c *client.Client, tip int64, lastN int64) (int64, error) {
	start := tip - lastN + 1
	if start < 0 {
		start = 0
	}
	var cnt int64
	for h := start; h <= tip; h++ {
		blk, err := c.GetBlockByNumber(ctx, h, false) // 仍然只传 ctx
		if err != nil {
			return 0, fmt.Errorf("get block %d: %w", h, err)
		}
		for _, anyHash := range blk.Transactions {
			hashStr, ok := anyHash.(string)
			if !ok {
				continue
			}
			rcp, err := c.GetTransactionReceipt(ctx, common.HexToHash(hashStr))
			if err != nil {
				return 0, fmt.Errorf("receipt %s: %w", hashStr, err)
			}
			if (rcp.ContractAddress != common.Address{}) { // 非零地址 => 合约创建
				cnt++
			}
		}
	}
	return cnt, nil
}
