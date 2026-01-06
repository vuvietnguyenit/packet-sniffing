package cmd

import (
	"os"

	"git.itim.vn/docker/packet-sniffer/pkg/cmd/mysql"
	"git.itim.vn/docker/packet-sniffer/pkg/cmd/redis"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "packet-sniffer",
}

func init() {
	rootCmd.AddCommand(mysql.MysqlErrorResponseCmd)
	rootCmd.AddCommand(redis.RedisErrorResponseCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
