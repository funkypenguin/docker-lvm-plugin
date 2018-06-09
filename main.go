package main

import (
	"flag"
	"fmt"
	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sirupsen/logrus"
	"os"
)

const (
	lvmHome              = "/var/lib/docker-lvm-plugin"
	lvmVolumesConfigPath = "/var/lib/docker-lvm-plugin/lvmVolumesConfig.json"
	lvmCountConfigPath   = "/var/lib/docker-lvm-plugin/lvmCountConfig.json"
	socket               = "/run/docker/plugins/lvm.sock"
)

var (
	flVersion  *bool
	flDebug    *bool
	flVgConfig *string
)

func init() {
	flVersion = flag.Bool("version", false, "Print version information and quit")
	flDebug = flag.Bool("debug", false, "Enable debug logging")
	flVgConfig = flag.String("vgConfig", "VOLUME_GROUP", "Name of the volume group environment variable.")
}

func main() {

	flag.Parse()

	if *flVersion {
		fmt.Fprint(os.Stdout, "docker lvm plugin version: 1.0.0\n")
		return
	}

	if *flDebug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.WithField("args", os.Args).Debug("arguments")
	}

	if _, err := os.Stat(lvmHome); err != nil {
		if !os.IsNotExist(err) {
			logrus.WithError(err).WithField("home", lvmHome).Fatal("cannot stat home")
		}
		logrus.WithField("home", lvmHome).Debug("Created home dir")
		if err := os.MkdirAll(lvmHome, 0700); err != nil {
			logrus.WithError(err).WithField("home", lvmHome).Fatal("cannot create home")
		}
	}

	lvm, err := newDriver(lvmHome, *flVgConfig)
	if err != nil {
		logrus.WithError(err).Fatal("error initializing lvmDriver")
	}

	if err := loadFromDisk(lvm); err != nil {
		logrus.WithError(err).Fatal("error restoring lvmDriver volume map")
	}

	h := volume.NewHandler(lvm)
	logrus.WithField("handler", h).Debug("new handler")
	if err := h.ServeUnix(socket, 0); err != nil {
		logrus.WithError(err).Fatal("cannot serve unix socket")
	}
}
