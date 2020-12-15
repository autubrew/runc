package vkernel

import (
	"bytes"
	"os"

	"github.com/opencontainers/runc/libcontainer/configs"
	"golang.org/x/sys/unix"
)

var (
	release string
)

// VKernel defines configuration information for current container's vkernel module.
type VKernel struct {
	Name string
	Path string // module path inside container
}

func getRelease() (string, error) {

	var u unix.Utsname
	if err := unix.Uname(&u); err != nil {
		return "", err
	}

	return string(u.Release[:bytes.IndexByte(u.Release[:], 0)]), nil
}

// New create a vkernel instance and initializes some global variables
func New() (vkn *VKernel, err error) {

	release, err = getRelease()
	if err != nil {
		return nil, err
	}

	vkn = &VKernel{
		Name: "vkernel",
		Path: "/lib/modules/" + release + "/extra/vkernel",
	}
	return vkn, nil
}

// ConfigureMountPath add vkernel module mount path to config
func (vkn *VKernel) ConfigureMountPath(mounts []*configs.Mount) ([]*configs.Mount, error) {

	dir := vkn.Path
	mounts = append(mounts, &configs.Mount{
		Source:      dir,
		Destination: dir,
		Device:      "bind",
		Flags:       unix.MS_BIND | unix.MS_REC,
		PremountCmds: []configs.Command{
			{Path: "touch", Args: []string{dir}},
		},
	})

	return mounts, nil
}

// InitVKernel initializes a vkernel module
func (vkn *VKernel) InitVKernel(params string, flags int) error {

	path := vkn.Path + "/" + vkn.Name + ".ko"
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return unix.FinitModule(int(f.Fd()), params, flags)

}

// DeleteVKernel uninstall the vkernel module
func (vkn *VKernel) DeleteVKernel() error {

	return unix.DeleteModule(vkn.Name, 0)

}
