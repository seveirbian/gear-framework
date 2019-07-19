package cmd

import (
    "github.com/seveirbian/gear/push"
    "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "path/filepath"
)

var pushUsage = `Usage:  gear push GEARIMAGENAME:TAG

Options:
  -m, --manager-ip          Manager node's ip address
  -p, --manager-port        Manager node's port(default 2019)
`

var (
    pushManagerIP string
    pushManagerPort string

    GearPath = "/var/lib/gear/"
    GearBuildPath = filepath.Join(GearPath, "build")
)

func init() {
    rootCmd.AddCommand(pushCmd)
    pushCmd.SetUsageTemplate(pushUsage)
    pushCmd.Flags().StringVarP(&pushManagerIP, "manager-ip", "m", "", "Manager node's ip address")
    pushCmd.MarkFlagRequired("manager-ip")
    pushCmd.Flags().StringVarP(&pushManagerPort, "manager-port", "p", "2019", "Manager node's port")
}

var pushCmd = &cobra.Command{
    Use:   "push",
    Short: "Push a gear image's files to backendstorage",
    Long:  `Push a gear image's files to backendstorage`,
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        gImageName, gImageTag := push.ParseImage(args[0])
        gFIlesDir := filepath.Join(GearBuildPath, gImageName+":"+gImageTag, "files")

        pusher, err := push.InitPusher(gFIlesDir, pushManagerIP, pushManagerPort, false)
        if err != nil {
            logrus.Fatal("Fail to init a pusher to push gear image...")
        }

        pusher.Push()
    },
}
