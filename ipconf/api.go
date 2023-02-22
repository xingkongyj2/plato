package ipconf

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hardcore-os/plato/ipconf/domain"
)

type Response struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

// GetIpInfoList API 适配应用层
func GetIpInfoList(c context.Context, ctx *app.RequestContext) {
	//捕捉异常。在这里不捕捉异常只会使本次调用错误，不会导致整个ip configÍ奔溃
	defer func() {
		if err := recover(); err != nil {
			ctx.JSON(consts.StatusBadRequest, utils.H{"err": err})
		}
	}()
	// Step0: 构建客户请求信息
	ipConfCtx := domain.BuildIpConfContext(&c, ctx)
	// Step1: 进行ip调度
	eds := domain.Dispatch(ipConfCtx)
	// Step2: 根据得分取top5返回
	ipConfCtx.AppCtx.JSON(consts.StatusOK, packRes(top5Endports(eds)))
}
