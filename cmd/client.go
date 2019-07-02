package cmd

import (
	"github.com/seveirbian/gear/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	clientUsage = `Usage:  gear client

Options:
  -m, --manager-ip          Manager node's ip address
  -p, --manager-port        Manager node's port(default 2019)
  -t, --monitor-ip          Monitor node's ip address
      --monitor-port        Monitor node's port(default 2021)
      --enable-p2p          Enable the clients to construct a p2p cluster
`
	managerIP string
	managerPort string
	clientMonitorIP string
	clientMonitorPort string
	enableP2p bool
)

func init() {
	rootCmd.AddCommand(clientCmd)
	clientCmd.SetUsageTemplate(clientUsage)
	clientCmd.Flags().StringVarP(&managerIP, "manager-ip", "m", "", "Manager node's ip address")
	clientCmd.MarkFlagRequired("manager-ip")
	clientCmd.Flags().StringVarP(&managerPort, "manager-port", "p", "2019", "Manager node's port")
	clientCmd.Flags().StringVarP(&clientMonitorIP, "monitor-ip", "t", "", "Monitor node's ip address")
	clientCmd.MarkFlagRequired("monitor-ip")
	clientCmd.Flags().StringVarP(&clientMonitorPort, "monitor-port", "", "2021", "Monitor node's port")
	
	clientCmd.Flags().BoolVarP(&enableP2p, "enable-p2p", "", false, "Enable p2p")
}

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a p2p cluster client",
	Long:  `Start a p2p cluster client`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cli, err := client.Init(managerIP, managerPort, clientMonitorIP, clientMonitorPort, enableP2p)
		if err != nil {
		    logrus.Fatal("Fail to init a client...")
		}

		cli.Start()
	},
}
