package subscan

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/itering/substrate-api-rpc"
	"github.com/itering/substrate-api-rpc/metadata"
	"github.com/itering/substrate-api-rpc/storage"
	"github.com/itering/substrate-api-rpc/storageKey"
	"github.com/itering/substrate-api-rpc/util"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"math/big"
	"strings"
	"time"

	"github.com/zhangxiaofeng05/subscan/internal/chainrpc"
	"github.com/zhangxiaofeng05/subscan/internal/utils"
	"github.com/zhangxiaofeng05/subscan/internal/utils/consts"
)

func Init(url string) *chainrpc.Client {
	rpcclient, err := chainrpc.DialForWallet(url)
	if err != nil {
		logrus.Fatalf("connect block chain node url err : %s", err)
	}
	return rpcclient
}

type ExtrinsicDetail struct {
	Params []ExtrinsicParam `json:"params"`
}

// GetBlock get block by height
func GetBlock(ctx context.Context, rpcclient *chainrpc.Client, height int64) *Block {
	data := make([]Transaction, 0)

	// block hash
	blockHash, err := rpcclient.ChainGetBlockHash(ctx, height)
	if err != nil {
		logrus.Fatal(err)
	}

	// block
	block, err := rpcclient.ChainGetBlock(ctx, blockHash)
	if err != nil {
		logrus.Fatal(err)
	}

	// event
	event, err := rpcclient.StateGetStorageAt(ctx, consts.EventStorageKey, blockHash)
	if err != nil {
		logrus.Fatal(err)
	}

	// runtime
	runtimeVersion, err := rpcclient.ChainGetRuntimeVersion(ctx, blockHash)
	if err != nil {
		logrus.Fatal(err)
	}
	if runtimeVersion == nil {
		logrus.Fatal("get runtimeVersion is nil")
	}
	specVersion := runtimeVersion.SpecVersion

	if block == nil || specVersion == -1 {
		logrus.Fatal("nil block data")
	}

	chainBlock := &ChainBlock{
		BlockNum: int(height),
		//BlockTimestamp:
		Hash:        blockHash,
		ParentHash:  block.Block.Header.ParentHash,
		StateRoot:   block.Block.Header.StateRoot,
		Extrinsics:  utils.ToString(block.Block.Extrinsics),
		Logs:        utils.ToString(block.Block.Header.Digest.Logs),
		Event:       event,
		SpecVersion: specVersion,
	}

	raw := &metadata.RuntimeRaw{
		Spec: specVersion,
	}
	getMetadata, err := rpcclient.StateGetMetadata(ctx, blockHash)
	if err != nil {
		logrus.Fatal(err)
	}
	raw.Raw = getMetadata

	metadataInstant := metadata.Process(raw)

	var (
		decodeEvent      interface{}
		encodeExtrinsics []string
		decodeExtrinsics []map[string]interface{}
	)

	err = json.Unmarshal([]byte(chainBlock.Extrinsics), &encodeExtrinsics)
	if err != nil {
		logrus.Fatal(err)
	}

	// Event
	decodeEvent, err = substrate.DecodeEvent(chainBlock.Event, metadataInstant, chainBlock.SpecVersion)
	if err != nil {
		logrus.Fatal(err)
	}

	// Extrinsic
	decodeExtrinsics, err = substrate.DecodeExtrinsic(encodeExtrinsics, metadataInstant, chainBlock.SpecVersion)
	if err != nil {
		logrus.Fatal(err)
	}

	// Log
	var rawList []string
	err = json.Unmarshal([]byte(chainBlock.Logs), &rawList)
	if err != nil {
		logrus.Fatal(err)
	}
	//decodeLogs, err := substrate.DecodeLogDigest(rawList)
	//if err != nil {
	//	logrus.Fatal(err)
	//}

	var chainEvent []ChainEvent
	utils.UnmarshalAny(&chainEvent, decodeEvent)
	eventMap := checkoutExtrinsicEvents(chainEvent, chainBlock.BlockNum)

	//extrinsicsCount, blockTimestamp, extrinsicHash, extrinsicFee, err := createExtrinsic(ctx, rpcclient, chainBlock, encodeExtrinsics, decodeExtrinsics, eventMap)
	//if err != nil {
	//	logrus.Fatal(err)
	//}

	//chainBlock.BlockTimestamp = blockTimestamp

	//eventCount, err := addEvent(chainBlock, e, extrinsicHash, extrinsicFee)
	//if err != nil {
	//	logrus.Fatal(err)
	//}

	//validator, err := EmitLog(chainBlock.BlockNum, decodeLogs, ValidatorsList(ctx, rpcclient, chainBlock.Hash))
	//if err != nil {
	//	logrus.Fatal(err)
	//}

	//chainBlock.EventCount = eventCount
	//chainBlock.ExtrinsicsCount = extrinsicsCount
	//chainBlock.BlockTimestamp = blockTimestamp

	// pause
	//_ = decodeLogs
	//_ = extrinsicsCount
	//_ = blockTimestamp
	//_ = extrinsicHash
	//_ = extrinsicFee
	//logrus.Fatal("pause")

	// block data
	var (
		blockTimestamp int
		e              []ChainExtrinsic
		//err            error
	)
	extrinsicFee := make(map[string]decimal.Decimal)

	eb, err := json.Marshal(decodeExtrinsics)
	if err != nil {
		logrus.Fatal(err)
	}
	err = json.Unmarshal(eb, &e)
	if err != nil {
		logrus.Fatal(err)
	}

	hash := make(map[string]string)
	for index, extrinsic := range e {
		extrinsic.CallModule = strings.ToLower(extrinsic.CallModule)
		extrinsic.ExtrinsicIndex = fmt.Sprintf("%d-%d", extrinsic.BlockNum, index)
		extrinsic.Success = getExtrinsicSuccess(eventMap[extrinsic.ExtrinsicIndex])

		if tp := getTimestamp(&extrinsic); tp > 0 {
			blockTimestamp = tp
		}
		extrinsic.BlockTimestamp = blockTimestamp
		if extrinsic.ExtrinsicHash != "" {

			fee, _ := getExtrinsicFee(ctx, rpcclient, encodeExtrinsics[index])
			extrinsic.Fee = fee

			extrinsicFee[extrinsic.ExtrinsicIndex] = fee
			hash[extrinsic.ExtrinsicIndex] = extrinsic.ExtrinsicHash
		}

		var ed ExtrinsicDetail
		utils.UnmarshalAny(&ed.Params, extrinsic.Params)

		// only get type: balances-transfer
		if extrinsic.CallModule == "balances" && extrinsic.CallModuleFunction == "transfer" {
			one := List{
				From:   extrinsic.AccountId,
				Failed: !extrinsic.Success,
			}
			for _, param := range ed.Params {
				if param.Name == "dest" && param.Type == "[U8; 20]" {
					one.To = param.Value.(string)
				}
				if param.Name == "value" && param.Type == "compact<U128>" {
					bi := big.Int{}
					value, ok := bi.SetString(param.Value.(string), 10)
					if ok {
						one.Value = value.Int64()
					}
				}
			}
			data = append(data, Transaction{
				Height:    block.Block.Header.Number.Int64(),
				Hash:      extrinsic.ExtrinsicHash,
				IsSuccess: extrinsic.Success,
				Fee:       extrinsicFee[extrinsic.ExtrinsicIndex].CoefficientInt64(),
				List:      []List{one},
				Timestamp: time.Unix(int64(blockTimestamp), 0),
				Executor:  extrinsic.AccountId,
			})
			if !extrinsic.Success {
				logrus.Fatalf("get target: extrinsic.BlockNum:%d extrinsic.ExtrinsicHash:%s", extrinsic.BlockNum, extrinsic.ExtrinsicHash)
			}
		}
	}

	res := &Block{
		Data: data,
	}

	fmt.Printf("block: %+v\n", res)

	return res
}

func checkoutExtrinsicEvents(e []ChainEvent, blockNumInt int) map[string][]ChainEvent {
	eventMap := make(map[string][]ChainEvent)
	for _, event := range e {
		extrinsicIndex := fmt.Sprintf("%d-%d", blockNumInt, event.ExtrinsicIdx)
		eventMap[extrinsicIndex] = append(eventMap[extrinsicIndex], event)
	}
	return eventMap
}

func createExtrinsic(ctx context.Context,
	rpcclient *chainrpc.Client,
	block *ChainBlock,
	encodeExtrinsics []string,
	decodeExtrinsics []map[string]interface{},
	eventMap map[string][]ChainEvent,
) (int, int, map[string]string, map[string]decimal.Decimal, error) {
	var (
		blockTimestamp int
		e              []ChainExtrinsic
		err            error
	)
	extrinsicFee := make(map[string]decimal.Decimal)

	eb, err := json.Marshal(decodeExtrinsics)
	if err != nil {
		logrus.Fatal(err)
	}
	err = json.Unmarshal(eb, &e)
	if err != nil {
		logrus.Fatal(err)
	}

	hash := make(map[string]string)

	for index, extrinsic := range e {
		extrinsic.CallModule = strings.ToLower(extrinsic.CallModule)
		extrinsic.BlockNum = block.BlockNum
		extrinsic.ExtrinsicIndex = fmt.Sprintf("%d-%d", extrinsic.BlockNum, index)
		extrinsic.Success = getExtrinsicSuccess(eventMap[extrinsic.ExtrinsicIndex])

		if tp := getTimestamp(&extrinsic); tp > 0 {
			blockTimestamp = tp
		}
		extrinsic.BlockTimestamp = blockTimestamp
		if extrinsic.ExtrinsicHash != "" {

			fee, _ := getExtrinsicFee(ctx, rpcclient, encodeExtrinsics[index])
			extrinsic.Fee = fee

			extrinsicFee[extrinsic.ExtrinsicIndex] = fee
			hash[extrinsic.ExtrinsicIndex] = extrinsic.ExtrinsicHash
		}

		if extrinsic.CallModule == "balances" && extrinsic.CallModuleFunction == "transfer" {
			logrus.Infof("extrinsic: %#v\n", extrinsic)
			logrus.Infof("extrinsic.Fee: %v", extrinsic.Fee.String())
			if !extrinsic.Success {
				logrus.Fatalf("get target: extrinsic.BlockNum:%d extrinsic.ExtrinsicHash:%s", extrinsic.BlockNum, extrinsic.ExtrinsicHash)
			}
		}

		//if err = s.dao.CreateExtrinsic(c, txn, &extrinsic); err == nil {
		//	go s.emitExtrinsic(block, &extrinsic, eventMap[extrinsic.ExtrinsicIndex])
		//} else {
		//	return 0, 0, nil, nil, err
		//}
	}
	return len(e), blockTimestamp, hash, extrinsicFee, err
}

// https://wiki.polkadot.network/docs/maintain-errors#polkadotjs-apps-explorer
// 例如：https://moonriver.subscan.io/extrinsic/3343805-4
func getExtrinsicSuccess(e []ChainEvent) bool {
	for _, event := range e {
		if strings.EqualFold(event.ModuleId, "System") && strings.EqualFold(event.EventId, "ExtrinsicFailed") {
			return false
		}
	}
	return true
}

func getTimestamp(extrinsic *ChainExtrinsic) (blockTimestamp int) {
	if extrinsic.CallModule != "timestamp" {
		return
	}

	var paramsInstant []ExtrinsicParam
	utils.UnmarshalAny(&paramsInstant, extrinsic.Params)

	for _, p := range paramsInstant {
		if p.Name == "now" {
			if strings.EqualFold(p.Type, "compact<U64>") {
				return int(utils.Int64FromInterface(p.Value) / 1000)
			}
			extrinsic.BlockTimestamp = utils.IntFromInterface(p.Value)
			return extrinsic.BlockTimestamp
		}
	}
	return
}

// getExtrinsicFee
func getExtrinsicFee(ctx context.Context, rpcclient *chainrpc.Client, encodeExtrinsic string) (fee decimal.Decimal, err error) {
	paymentInfo, err := rpcclient.PaymentQueryInfo(ctx, encodeExtrinsic)
	if err != nil {
		logrus.Fatal(err)
	}
	if paymentInfo != nil {
		return paymentInfo.PartialFee, nil
	}
	return decimal.Zero, err
}

func addEvent(
	block *ChainBlock,
	e []ChainEvent,
	hashMap map[string]string,
	feeMap map[string]decimal.Decimal) (eventCount int, err error) {

	for _, event := range e {
		event.ModuleId = strings.ToLower(event.ModuleId)
		event.ExtrinsicHash = hashMap[fmt.Sprintf("%d-%d", block.BlockNum, event.ExtrinsicIdx)]
		event.EventIndex = fmt.Sprintf("%d-%d", block.BlockNum, event.ExtrinsicIdx)
		event.BlockNum = block.BlockNum

		//if err = s.dao.CreateEvent(txn, &event); err == nil {
		//	go s.emitEvent(block, &event, feeMap[event.EventIndex])
		//} else {
		//	return 0, err
		//}
		eventCount++
	}
	return eventCount, err
}

func emitLog(blockNum int, l []storage.DecoderLog, validatorList []string) (validator string, err error) {
	for _, logData := range l {
		dataStr := utils.ToString(logData.Value)

		//ce := model.ChainLog{
		//	LogIndex:  fmt.Sprintf("%d-%d", blockNum, index),
		//	BlockNum:  blockNum,
		//	LogType:   logData.Type,
		//	Data:      dataStr,
		//	Finalized: finalized,
		//}
		//if err = s.dao.CreateLog(txn, &ce); err != nil {
		//	return "", err
		//}

		// check validator
		if strings.EqualFold(logData.Type, "PreRuntime") {
			validator = substrate.ExtractAuthor([]byte(dataStr), validatorList)
		}

	}
	return validator, err
}

func validatorsList(ctx context.Context, rpcclient *chainrpc.Client, blockHash string) (validatorList []string) {
	validatorsRaw, err := readStorage(ctx, rpcclient, "Session", "Validators", blockHash)
	if err != nil {
		logrus.Fatal(err)
	}
	for _, addr := range validatorsRaw.ToStringSlice() {
		validatorList = append(validatorList, util.TrimHex(addr))
	}
	return
}

func readStorage(ctx context.Context,
	rpcclient *chainrpc.Client,
	module, prefix string,
	blockHash string,
	arg ...string,
) (r storage.StateStorage, err error) {
	key := storageKey.EncodeStorageKey(module, prefix, arg...)
	logrus.Infof("key.EncodeKey: %s blockHash: %s", key.EncodeKey, blockHash)
	dataHex, err := rpcclient.StateGetStorageAt(ctx, utils.AddHex(key.EncodeKey), blockHash)
	if err != nil {
		logrus.Fatal(err)
	}
	return storage.Decode(dataHex, key.ScaleType, nil)
}
