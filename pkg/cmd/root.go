package cmd

import (
	"os"

	"git.itim.vn/docker/redis-error-sniffer/pkg/cmd/mysql"
	"git.itim.vn/docker/redis-error-sniffer/pkg/cmd/redis"
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
