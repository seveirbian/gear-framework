package cmd

import (
    "github.com/seveirbian/gear/watcher"
    "github.com/spf13/cobra"
)

var watcherUsage = `Usage:  gear watcher -i IP -p PORT

Options:
  -i, --server-ip          Server node's ip address
  -p, --server-port        Server node's port(default 2021)`

var (
    watcherSIp string
    watcherSPort string
)

func init() {
    rootCmd.AddCommand(watcherCmd)
    watcherCmd.SetUsageTemplate(watcherUsage)
    watcherCmd.Flags().StringVarP(&watcherSIp, "server-ip", "i", "", "Server node's ip address")
    watcherCmd.MarkFlagRequired("server-ip")
    watcherCmd.Flags().StringVarP(&watcherSPort, "server-port", "p", "2021", "Server node's port")
}

var watcherCmd = &cobra.Command{
    Use:   "watcher",
    Short: "Watch the accessed file in images",
    Long:  `Watch the accessed file in images`,
    Args:  cobra.NoArgs,
    Run: func(cmd *cobra.Command, args []string) {
        watcher.Watch("/", watcher.Server{Ip:watcherSIp, Port:watcherSPort})
    },
}
