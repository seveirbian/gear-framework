package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "gear",
	Short: "Gear is a fast docker container deployment system",
	Long: `A fast docker container deployment system.
Complete documentation is available at https://github.com/seveirbian/gear`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Fatal("Fail execute rootCmd")
		os.Exit(1)
	}
}
