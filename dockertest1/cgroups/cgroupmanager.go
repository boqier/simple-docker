package cgroups

import (
	"dockertest1/cgroups/fs1"
)

type cgroupManager struct {
	Path           string
	ResourceConfig *fs1.ResourceConfig
}

func NewCgroupManager(path string) *cgroupManager {
	return &cgroupManager{
		Path: path,
	}
}

func (c *cgroupManager) Apply(pid int) error {
	for _, Subsystemsints := range fs1.Subsystemsints {
		Subsystemsints.Apply(c.Path, pid)
	}
	return nil
}
func (c *cgroupManager) Set(res *fs1.ResourceConfig) error {
	for _, Subsystemsints := range fs1.Subsystemsints {
		if err := Subsystemsints.Set(c.Path, res); err != nil {
			return err
		}
	}
	return nil

}
func (c *cgroupManager) Remove() error {
	for _, Subsystemsints := range fs1.Subsystemsints {
		if err := Subsystemsints.Remove(c.Path); err != nil {
			return err
		}
	}
	return nil
}
