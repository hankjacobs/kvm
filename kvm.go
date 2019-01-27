package kvm

import (
	"os"
	"syscall"
)

var osIoctl ioctler = &osIoctler{}

// VM is a KVM VM
type VM struct {
	Fd       uintptr
	MMapSize int
}

// CreateVM creates a new VM
func CreateVM() (*VM, error) {
	file, err := os.OpenFile("/dev/kvm", os.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	kvmfd := file.Fd()
	vmfd, err := osIoctl.ioctl(kvmfd, ioctlKVMCreateVM, 0)
	if err != nil {
		return nil, err
	}

	mmapsize, err := osIoctl.ioctl(kvmfd, ioctlKVMGetVCPUMMAPSize, 0)
	if err != nil {
		return nil, err
	}

	vm := VM{
		Fd:       vmfd,
		MMapSize: int(mmapsize),
	}

	return &vm, nil
}
