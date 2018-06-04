package main

import (
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"github.com/sirupsen/logrus"
)

type lvmDriver struct {
	home     string
	vgConfig string
	volumes  map[string]*vol
	count    map[string]int
	mu       sync.RWMutex
}

type vol struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountpoint"`
	Type       string `json:"type"`
	Source     string `json:"source"`
	KeyFile    string `json:"keyfile"`
}

func newDriver(home, vgConfig string) (*lvmDriver, error) {
	return &lvmDriver{
		home:     home,
		vgConfig: vgConfig,
		volumes:  make(map[string]*vol),
		count:    make(map[string]int),
	}, nil
}

func (l *lvmDriver) Create(req *volume.CreateRequest) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.volumes[req.Name]; exists {
		return nil
	}

	vgName, err := getVolumegroupName(l.vgConfig)
	if err != nil {
		return err
	}

	keyFile, ok := req.Options["keyfile"]
	hasKeyFile := ok && keyFile != ""
	if hasKeyFile {
		if err := keyFileExists(keyFile); err != nil {
			return err
		}
		if err := cryptsetupInstalled(); err != nil {
			return err
		}
	}

	cmdArgs := []string{"-n", req.Name, "--setactivationskip", "n"}
	snap, ok := req.Options["snapshot"]
	isSnapshot := ok && snap != ""
	isThinSnap := false
	if isSnapshot {
		if hasKeyFile {
			return fmt.Errorf("please don't specify --opt keyfile= for snapshots")
		}
		if isThinSnap, err = isThinlyProvisioned(vgName, snap); err != nil {
			logrus.WithError(err).Error("Create: lvdisplayGrep error")
			return fmt.Errorf("error creating volume from snapshot")
		}
	}
	s, ok := req.Options["size"]
	hasSize := ok && s != ""

	if !hasSize && !isThinSnap {
		return fmt.Errorf("please specify a size with --opt size=")
	}

	if hasSize && isThinSnap {
		return fmt.Errorf("please don't specify --opt size= for thin snapshots")
	}

	if isSnapshot {
		cmdArgs = append(cmdArgs, "--snapshot")
		if hasSize {
			cmdArgs = append(cmdArgs, "--size", s)
		}
		cmdArgs = append(cmdArgs, vgName+"/"+snap)
	} else if thin, ok := req.Options["thinpool"]; ok && thin != "" {
		cmdArgs = append(cmdArgs, "--virtualsize", s)
		cmdArgs = append(cmdArgs, "--thin")
		cmdArgs = append(cmdArgs, vgName+"/"+thin)
	} else {
		cmdArgs = append(cmdArgs, "--size", s)
		cmdArgs = append(cmdArgs, vgName)
	}
	cmd := exec.Command("lvcreate", cmdArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		logrus.WithError(err).WithField("out", string(out)).Errorf("create: lvcreate error")
		return fmt.Errorf("error creating volume: %s", string(out))
	}

	defer func() {
		if err != nil {
			removeLogicalVolume(req.Name, vgName)
		}
	}()

	if !isSnapshot {
		device := logicalDevice(vgName, req.Name)

		if hasKeyFile {
			cmd = exec.Command("cryptsetup", "-q", "-d", keyFile, "luksFormat", device)
			if out, err := cmd.CombinedOutput(); err != nil {
				logrus.WithError(err).WithField("out", string(out)).Error("create: cryptsetup error")
				return fmt.Errorf("error encrypting volume")
			}

			if out, err := luksOpen(vgName, req.Name, keyFile); err != nil {
				logrus.WithError(err).WithField("out", string(out)).Error("create: cryptsetup error")
				return fmt.Errorf("error opening encrypted volume")
			}

			defer func() {
				if err != nil {
					if out, err := luksClose(req.Name); err != nil {
						logrus.WithError(err).WithField("out", string(out)).Error("create: cryptsetup error")
					}
				}
			}()

			device = luksDevice(req.Name)
		}

		cmd = exec.Command("mkfs.xfs", device)
		if out, err := cmd.CombinedOutput(); err != nil {
			logrus.WithError(err).WithField("out", string(out)).Error("Create: mkfs.xfs error")
			return fmt.Errorf("error partitioning volume")
		}

		if hasKeyFile {
			if out, err := luksClose(req.Name); err != nil {
				logrus.WithError(err).WithField("out", string(out)).Error("Create: cryptsetup error")
				return fmt.Errorf("error closing encrypted volume")
			}
		}
	}

	mp := getMountpoint(l.home, req.Name)
	err = os.MkdirAll(mp, 0700)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			os.RemoveAll(mp)
		}
	}()

	v := &vol{Name: req.Name, MountPoint: mp}
	if isSnapshot {
		source := l.volumes[snap]
		v.Type = "Snapshot"
		v.Source = snap
		v.KeyFile = source.KeyFile
	} else if hasKeyFile {
		v.KeyFile = keyFile
	}
	l.volumes[v.Name] = v
	l.count[v.Name] = 0
	err = saveToDisk(l.volumes, l.count)
	if err != nil {
		return err
	}
	return nil
}

func (l *lvmDriver) List() (*volume.ListResponse, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	var ls []*volume.Volume
	for _, vol := range l.volumes {
		v := &volume.Volume{
			Name:       vol.Name,
			Mountpoint: vol.MountPoint,
			// TO-DO: Find the significance of status field, and add that to volume.Volume
		}
		ls = append(ls, v)
	}
	return &volume.ListResponse{Volumes: ls}, nil
}

func (l *lvmDriver) Get(req *volume.GetRequest) (*volume.GetResponse, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	v, exists := l.volumes[req.Name]
	if !exists {
		return &volume.GetResponse{}, fmt.Errorf("no such volume")
	}
	var res volume.GetResponse
	res.Volume = &volume.Volume{
		Name:       v.Name,
		Mountpoint: v.MountPoint,
	}
	return &res, nil
}

func (l *lvmDriver) Remove(req *volume.RemoveRequest) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	vgName, err := getVolumegroupName(l.vgConfig)
	if err != nil {
		return err
	}

	isOrigin := func() bool {
		for _, vol := range l.volumes {
			if vol.Name == req.Name {
				continue
			}
			if vol.Type == "Snapshot" && vol.Source == req.Name {
				return true
			}
		}
		return false
	}()

	if isOrigin {
		return fmt.Errorf("error removing volume, all snapshot destinations must be removed before removing the original volume")
	}

	if err := os.RemoveAll(getMountpoint(l.home, req.Name)); err != nil {
		return err
	}

	if out, err := removeLogicalVolume(req.Name, vgName); err != nil {
		logrus.WithError(err).WithField("out", string(out)).Error("remove: removeLogicalVolume error")
		return fmt.Errorf("error removing volume")
	}

	delete(l.count, req.Name)
	delete(l.volumes, req.Name)
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return err
	}
	return nil
}

func (l *lvmDriver) Path(req *volume.PathRequest) (*volume.PathResponse, error) {
	return &volume.PathResponse{Mountpoint: getMountpoint(l.home, req.Name)}, nil
}

func (l *lvmDriver) Mount(req *volume.MountRequest) (*volume.MountResponse, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	vgName, err := getVolumegroupName(l.vgConfig)
	if err != nil {
		return &volume.MountResponse{}, err
	}

	isSnap, keyFile := func() (bool, string) {
		if v, ok := l.volumes[req.Name]; ok {
			if v.Type == "Snapshot" {
				return true, v.KeyFile
			}
			return false, v.KeyFile
		}
		return false, ""
	}()

	if l.count[req.Name] == 0 {
		device := logicalDevice(vgName, req.Name)

		if keyFile != "" {
			if err := keyFileExists(keyFile); err != nil {
				logrus.WithError(err).Error("mount")
				return &volume.MountResponse{}, err
			}
			if err := cryptsetupInstalled(); err != nil {
				logrus.WithError(err).Error("mount")
				return &volume.MountResponse{}, err
			}
			if out, err := luksOpen(vgName, req.Name, keyFile); err != nil {
				logrus.WithError(err).WithField("out", string(out)).Error("Mount: cryptsetup error")
				return &volume.MountResponse{}, fmt.Errorf("error opening encrypted volume")
			}
			defer func() {
				if err != nil {
					if out, err := luksClose(req.Name); err != nil {
						logrus.WithError(err).WithField("out", string(out)).Error("Mount: cryptsetup error")
					}
				}
			}()
			device = luksDevice(req.Name)
		}

		mountArgs := []string{device, getMountpoint(l.home, req.Name)}
		if isSnap {
			mountArgs = append([]string{"-o", "nouuid"}, mountArgs...)
		}
		cmd := exec.Command("mount", mountArgs...)
		if out, err := cmd.CombinedOutput(); err != nil {
			logrus.WithError(err).WithField("out", string(out)).Error("mount: mount error")
			return &volume.MountResponse{}, fmt.Errorf("error mouting volume")
		}
	}
	l.count[req.Name]++
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return &volume.MountResponse{}, err
	}
	return &volume.MountResponse{Mountpoint: getMountpoint(l.home, req.Name)}, nil
}

func (l *lvmDriver) Unmount(req *volume.UnmountRequest) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.count[req.Name] == 1 {
		cmd := exec.Command("umount", getMountpoint(l.home, req.Name))
		if out, err := cmd.CombinedOutput(); err != nil {
			logrus.WithError(err).WithField("out", string(out)).Error("unmount: unmount error")
			return fmt.Errorf("error unmounting volume")
		}
		if v, ok := l.volumes[req.Name]; ok && v.KeyFile != "" {
			if err := cryptsetupInstalled(); err != nil {
				logrus.WithError(err).Error("unmount")
				return err
			}
			if out, err := luksClose(req.Name); err != nil {
				logrus.WithError(err).WithField("out", string(out)).Error("unmount: cryptsetup error")
				return fmt.Errorf("error closing encrypted volume")
			}
		}
	}
	l.count[req.Name]--
	if err := saveToDisk(l.volumes, l.count); err != nil {
		return err
	}
	return nil
}

func (l *lvmDriver) Capabilities() *volume.CapabilitiesResponse {
	var res volume.CapabilitiesResponse
	res.Capabilities = volume.Capability{Scope: "local"}
	return &res
}

func luksOpen(vgName, volName, keyFile string) ([]byte, error) {
	cmd := exec.Command("cryptsetup", "-d", keyFile, "luksOpen", logicalDevice(vgName, volName), luksDeviceName(volName))
	if out, err := cmd.CombinedOutput(); err != nil {
		return out, err
	}
	return nil, nil
}

func luksClose(volName string) ([]byte, error) {
	cmd := exec.Command("cryptsetup", "luksClose", luksDeviceName(volName))
	if out, err := cmd.CombinedOutput(); err != nil {
		return out, err
	}
	return nil, nil
}
