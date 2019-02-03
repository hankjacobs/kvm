package kvm

import (
	"os"
	"syscall"
	"unsafe"
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

const (
	// UserMemoryFlagLogDirtyPages maps to KVM_MEM_LOG_DIRTY_PAGES
	UserMemoryFlagLogDirtyPages = 1 << iota

	// UserMemoryReadOnly maps to KVM_MEM_READONLY
	UserMemoryReadOnly
)

type kvmUserspaceMemoryRegion struct {
	Slot          uint32
	Flags         uint32
	GuestPhysAddr uint64
	MemorySize    uint64
	UserspaceAddr uint64
}

// MapUserMemory maps memory to the specified slot with the given flags
func (vm *VM) MapUserMemory(slot uint32, flags uint32, guestAddress uint64, memory []byte) error {
	userspaceAddr := unsafe.Pointer(&memory[0])
	region := kvmUserspaceMemoryRegion{
		Slot:          slot,
		Flags:         flags,
		GuestPhysAddr: guestAddress,
		MemorySize:    uint64(len(memory)),
		UserspaceAddr: uint64(uintptr(userspaceAddr)),
	}

	_, err := osIoctl.ioctl(vm.Fd, ioctlKVMSetUserMemoryRegion, uintptr(unsafe.Pointer(&region)))
	return err
}
