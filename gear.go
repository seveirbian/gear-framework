package main

import (
	"os"
	"path/filepath"

	"github.com/seveirbian/gear/cmd"
	"github.com/sirupsen/logrus"
)

var (
	GearPath             = "/var/lib/gear/"
	GearPrivateCachePath = filepath.Join(GearPath, "private")
	GearPublicCachePath  = filepath.Join(GearPath, "public")
	GearNFSPath          = filepath.Join(GearPath, "nfs")
	GearBuildPath        = filepath.Join(GearPath, "build")
	GearImagesPath       = filepath.Join(GearPath, "images")
	GearContainersPath   = filepath.Join(GearPath, "containers")
)

var (
	logger = logrus.WithField("gear", "init")
)

func init() {
	// create gear's home dir, if not exists, create one
	_, err := os.Stat(GearPath)
	if err != nil {
		err = os.MkdirAll(GearPath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearPath: %v \n", err)
		}
	}
	// create gear's private cache dir, if not exists, create one
	_, err = os.Stat(GearPrivateCachePath)
	if err != nil {
		err = os.MkdirAll(GearPrivateCachePath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearPrivateCachePath: %v \n", err)
		}
	}
	// create gear's public cache dir, if not exists, create one
	_, err = os.Stat(GearPublicCachePath)
	if err != nil {
		err = os.MkdirAll(GearPublicCachePath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearPublicCachePath: %v \n", err)
		}
	}
	// create gear's nfs dir, if not exists, create one
	_, err = os.Stat(GearNFSPath)
	if err != nil {
		err = os.MkdirAll(GearNFSPath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearNFSPath: %v \n", err)
		}
	}
	// create gear's build dir, if not exists, create one
	_, err = os.Stat(GearBuildPath)
	if err != nil {
		err = os.MkdirAll(GearBuildPath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearBuildPath: %v \n", err)
		}
	}
	// create gear's images dir, if not exists, create one
	_, err = os.Stat(GearImagesPath)
	if err != nil {
		err = os.MkdirAll(GearImagesPath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearImagesPath: %v \n", err)
		}
	}
	// create gear's contaienrs dir, if not exists, create one
	_, err = os.Stat(GearContainersPath)
	if err != nil {
		err = os.MkdirAll(GearContainersPath, os.ModePerm)
		if err != nil {
			logger.Debugf("Fail to create GearContainersPath: %v \n", err)
		}
	}
}

func main() {
	cmd.Execute()
}
