package cmd

import (
	"github.com/hardcore-os/plato/client"
	"github.com/spf13/cobra"
)

// 包加载的时候会先执行init函数
func init() {
	rootCmd.AddCommand(clientCmd)
}

// 定义一个子命令client，ClientHandle是回掉函数
var clientCmd = &cobra.Command{
	Use: "client",
	Run: ClientHandle,
}

func ClientHandle(cmd *cobra.Command, args []string) {
	client.RunMain()
}
