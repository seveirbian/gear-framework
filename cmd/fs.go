package cmd

import (
	"github.com/seveirbian/gear/fs"
	"github.com/spf13/cobra"
)

var fsUsage = `Usage:  gear fs -i IndexImagePath -p PrivateCachePath MountPoint
IndexImagePath, PrivateCachePath and MountPoint must be absolute path
`

var (
	IndexImagePath string
	PrivateCachePath string
)

func init() {
	rootCmd.AddCommand(fsCmd)
	fsCmd.SetUsageTemplate(fsUsage)
	fsCmd.Flags().StringVarP(&IndexImagePath, "indexImagePath", "i", "", "Index image path")
	fsCmd.MarkFlagRequired("indexImagePath")
	fsCmd.Flags().StringVarP(&PrivateCachePath, "privateCachePath", "p", "", "Private cache path")
	fsCmd.MarkFlagRequired("privateCachePath")
}

var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "Create a download-on-demand fs from a gear image directory",
	Long:  `Create a download-on-demand fs from a gear image directory`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		gearFS := &fs.GearFS {
			MountPoint: args[0], 
			IndexImagePath: IndexImagePath, 
			PrivateCachePath: PrivateCachePath, 
		}

		gearFS.Start()
	},
}
