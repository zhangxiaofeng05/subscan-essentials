package main

import (
	"context"
	"github.com/zhangxiaofeng05/subscan/subscan"
)

var (
	moonriverUrl = "https://rpc.api.moonriver.moonbeam.network"
)

func main() {
	rpcclient := subscan.Init(moonriverUrl)
	defer rpcclient.Close()

	ctx := context.Background()
	var height int64 = 3241439
	//var height int64 = 3192740 //3192754-7 Failed
	for {
		block := subscan.GetBlock(ctx, rpcclient, height)
		_ = block
		//height++
		break

	}

}
