package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ctx = context.Background()
	// url 			= "https://mainnet.infura.io/v3/4fce2c82228340669f3ffc9c5a6c7768" //for Infura
	// url         	= "http://127.0.0.1:8545" //for Ganache
	url         = "wss://mainnet.infura.io/ws/v3/4fce2c82228340669f3ffc9c5a6c7768" //for Infura websocket endpoint
	client, err = ethclient.Dial(url)
)

func currentBlock() {
	block, err := client.BlockByNumber(ctx, nil)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Current Block Number: ", block.Number())
	fmt.Println()
}

func createWallet() (string, string) {
	fmt.Println("==================START CREATE WALLET===================")

	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Println(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	fmt.Println("SAVE BUT DO NOT SHARE THIS (Private Key):", hexutil.Encode(privateKeyBytes))
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println("Public Key:", hexutil.Encode(publicKeyBytes))

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Address:", address)
	fmt.Println("==================END CREATE WALLET===================")

	fmt.Println()
	fmt.Println()

	return address, hexutil.Encode(publicKeyBytes)
}

func transfer_eth(fromPrivateKeyString, toAddressString string) {

	fmt.Println("==================START TRANSFER===================")

	fromPrivateKey, err := crypto.HexToECDSA(fromPrivateKeyString)
	if err != nil {
		log.Fatal(err)
	}
	fromPublicKey := fromPrivateKey.Public()
	fromPublicKeyECDSA, ok := fromPublicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: [From PublicKey] is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*fromPublicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal(err)
	}
	value := big.NewInt(1000000000000000000) // in wei (1 eth)
	gasLimit := uint64(21000)                // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	toAddress := common.HexToAddress(toAddressString)

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivateKey)
	if err != nil {
		log.Fatal(err)
	}
	err = client.SendTransaction(context.Background(), signedTx)

	if err != nil {
		fmt.Println("====sendTransaction Error:====", err)

	}
	fmt.Println("tx sent:", tx.Hash().Hex())

	fromBalance, err := client.BalanceAt(context.Background(), fromAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("From Balance", fromBalance)

	toBalance, err := client.BalanceAt(context.Background(), toAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("To Balance", toBalance)

	fmt.Println("==================END TRANSFER TEST===================")
	fmt.Println()
	fmt.Println()

}

func main() {
	currentBlock()

	//===================START TRANSFER=================
	// fromPrivateKeyString := "d64be871edc9fb0bd80a8ce0949c804fa37ef92c32f7333d76b7d1a672fecc5f"
	// toAddressString := "0xE2DcC343CA1FdDB2C8CE2B79E0dD6744A5f582C0"
	// transfer_eth(fromPrivateKeyString, toAddressString)
	//===================END TRANSFER=================

	pendingTxs := make(chan []byte)
	subscription, err := client.Client().EthSubscribe(context.Background(), pendingTxs, "newPendingTransactions")
	if err != nil {
		log.Fatal(err)
	}
	defer subscription.Unsubscribe()

	for {
		select {
		case err := <-subscription.Err():
			log.Fatal(err)
		case tx := <-pendingTxs:
			fmt.Println("New Pending Transaction:", tx)
		}
	}

	//===================SUBSCRIBE EVENT LOG=================
	// logs := make(chan types.Log)
	// query := ethereum.FilterQuery{}
	// sub, err := client.SubscribeFilterLogs(ctx, query, logs)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer sub.Unsubscribe()
	// for {
	// 	select {
	// 	case err := <-sub.Err():
	// 		log.Fatal(err)
	// 	case vLog := <-logs:
	// 		block, err := client.BlockByHash(context.Background(), vLog.BlockHash)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Println(block.TxHash())
	// 		fmt.Println(block.Hash().Hex())
	// 		fmt.Println(block.Number().Uint64())
	// 		fmt.Println(block.Time())
	// 		fmt.Println(block.Nonce())
	// 		fmt.Println(len(block.Transactions()))
	// 		pendingTransactionCount, err := client.PendingTransactionCount(context.Background())
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Println("pendingTransactionCount:", pendingTransactionCount)
	// 	}
	// }
	//===================END SUBSCRIBE EVENT LOG=================

	//===============START SUBSCRIBE NEW HEAD====================
	// headers := make(chan *types.Header)
	// sub, err := client.SubscribeNewHead(context.Background(), headers)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// for {
	// 	select {
	// 	case err := <-sub.Err():
	// 		log.Fatal(err)
	// 	case header := <-headers:
	// 		block, err := client.BlockByHash(context.Background(), header.Hash())
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		fmt.Println(block.Hash().Hex())
	// 		fmt.Println(block.Number().Uint64())
	// 		fmt.Println(block.Time())
	// 		fmt.Println(block.Nonce())
	// 		fmt.Println(len(block.Transactions()))
	// 	}
	// }
	//===============END SUBSCRIBE NEW HEAD====================

}
