package domain

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// 封装了3种Conext
type IpConfConext struct {
	//原始Hertz的Conext
	Ctx       *context.Context
	AppCtx    *app.RequestContext
	ClinetCtx *ClientConext
}

type ClientConext struct {
	IP string `json:"ip"`
}

func BuildIpConfContext(c *context.Context, ctx *app.RequestContext) *IpConfConext {
	ipConfConext := &IpConfConext{
		Ctx:       c,
		AppCtx:    ctx,
		ClinetCtx: &ClientConext{},
	}
	return ipConfConext
}
