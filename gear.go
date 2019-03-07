package main

import (
    "os"
    "path/filepath"

    "github.com/sirupsen/logrus"
    "github.com/seveirbian/gear/cmd"
)

// gear's home dir is /home/gear/
var GearRootPath = filepath.Join(os.Getenv("HOME"), "gear")

func init() {
    // create gear's home dir, if not exists, create one
    _, err := os.Stat(GearRootPath)
    if err != nil {
        err = os.MkdirAll(GearRootPath, os.ModePerm)
        if err != nil {
            logrus.WithFields(logrus.Fields{
                "err": err, 
                }).Fatal("Fail to create Gear's root dir")
        }
    }
}

func main() {

    cmd.Execute()
}