package cmd

import (
    "github.com/seveirbian/gear/monitor"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var monitorUsage = `Usage:  gear monitor REGISTRYIP:REGISTRYPORT

Options:
  -i, --manager-ip          Manager node's ip
      --manager-port        Manager node's port
  -n, --no-cleanup          Do not clean up building dirs
`

var (
    monitorManagerIp string
    monitorManagerPort string
    monitorNoClean bool
)

func init() {
    rootCmd.AddCommand(monitorCmd)
    monitorCmd.SetUsageTemplate(monitorUsage)
    monitorCmd.Flags().StringVarP(&monitorManagerIp, "manager-ip", "i", "", "Manager node's ip")
    monitorCmd.MarkFlagRequired("manager-ip")
    monitorCmd.Flags().StringVarP(&monitorManagerPort, "manager-port", "", "2019", "Manager node's port")
    monitorCmd.Flags().BoolVarP(&monitorNoClean, "no-cleanup", "n", false, "Do not clean up building dirs")
}

var monitorCmd = &cobra.Command{
    Use:   "monitor",
    Short: "Monitor a docker registry and build gear images for each docker image",
    Long:  `Monitor a docker registry and build gear images for each docker image`,
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        monitor, err := monitor.InitMonitor(args[0], monitorManagerIp, monitorManagerPort, monitorNoClean)
        if err != nil {
            logrus.Fatal("Fail to init a monitor to monitor docker reistry...")
        }

        err = monitor.Monitor()
        if err != nil {
            logrus.Fatal("Fail to monitor docker registry...")
        }
    },
}
