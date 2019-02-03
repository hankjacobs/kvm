package kvm

import (
	"syscall"

	"golang.org/x/sys/unix"
)

// KVM Ioctls
const (
	ioctlKVMGetAPIVersion          = 0xAE00
	ioctlKVMCreateVM               = 0xAE01
	ioctlKVMGetMSRIndexList        = 0xC004AE02
	ioctlKVMS390EnableSIE          = 0xAE06
	ioctlKVMCheckExtension         = 0xAE03
	ioctlKVMGetVCPUMMAPSize        = 0xAE04
	ioctlKVMGetSupportedCPUID      = 0xC008AE05
	ioctlKVMGetEmulatedCPUID       = 0xC008AE09
	ioctlKVMGetMSRFeatureIndexList = 0xC004AE0A
	ioctlKVMSetUserMemoryRegion    = 0x4020AE46
	ioctlKVMCreateVCPU             = 0xAE41
	ioctlKVMGetRegs                = 0x8090AE81
	ioctlKVMSetRegs                = 0x4090ae82
	ioctlKVMGetSRegs               = 0x8138AE83
	ioctlKVMSetSRegs               = 0x4138AE84
	ioctlKVMRun                    = 0xAE80
)

// https://github.com/golang/sys/blob/master/unix/syscall_unix.go#L33
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	// are these cases even relevant for KVM?
	case unix.EAGAIN:
		return syscall.EAGAIN
	case unix.EINVAL:
		return syscall.EINVAL
	case unix.ENOENT:
		return syscall.ENOENT
	case unix.EINTR:
		return syscall.EINTR
	}
	return e
}

// ioctler is an interface capable of calling ioctl
type ioctler interface {
	ioctl(fd uintptr, req uint, arg uintptr) (ret uintptr, err error)
}

// An osIoctler struct does OS calls to ioctl.
type osIoctler struct{}

func (osIoctler) ioctl(fd uintptr, req uint, arg uintptr) (ret uintptr, err error) {
	ret, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(fd),
		uintptr(req),
		arg,
	)
	if errno != 0 {
		err = errnoErr(errno)
	}

	return
}
