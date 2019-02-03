package kvm

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

func (vm *VM) CreateVCPU() (*VCPU, error) {
	fd, err := osIoctl.ioctl(vm.Fd, ioctlKVMCreateVCPU, 0)
	if err != nil {
		return nil, err
	}

	runMap, err := syscall.Mmap(int(fd), 0, vm.MMapSize, syscall.PROT_READ|syscall.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		syscall.Close(int(fd))
		return nil, err
	}

	run := (*runData)(unsafe.Pointer(&runMap[0]))

	return &VCPU{Fd: fd, run: run}, nil
}

type VCPU struct {
	Fd  uintptr
	run *runData
}

type Registers struct {
	Rax, Rbx, Rcx, Rdx uint64
	Rsi, Rdi, Rsp, Rbp uint64
	R8, R9, R10, R11   uint64
	R12, R13, R14, R15 uint64
	Rip, Rflags        uint64
}

func (c *VCPU) GetRegisters() (Registers, error) {
	regs := Registers{}
	_, err := osIoctl.ioctl(c.Fd, ioctlKVMGetRegs, uintptr(unsafe.Pointer(&regs)))
	return regs, err
}

func (c *VCPU) SetRegisters(regs Registers) error {
	_, err := osIoctl.ioctl(c.Fd, ioctlKVMSetRegs, uintptr(unsafe.Pointer(&regs)))
	return err
}

type Segment struct {
	Base                           uint64
	Limit                          uint32
	Selector                       uint16
	Type                           uint8
	Present, Dpl, Db, S, L, G, Avl uint8
	unusable                       uint8
	padding                        uint8
}

type Dtable struct {
	Base    uint64
	Limit   uint16
	padding [3]uint16
}

const NRInterrupts = 256

type SRegisters struct {
	Cs, Ds, Es, Fs, Gs, Ss  Segment
	Tr, Ldt                 Segment
	Gdt, Idt                Dtable
	Cr0, Cr2, Cr3, Cr4, Cr8 uint64
	Efer                    uint64
	ApicBase                uint64
	InterruptBitmap         [(NRInterrupts + 63) / 64]uint64
}

func (c *VCPU) GetSRegisters() (SRegisters, error) {
	sregs := SRegisters{}
	_, err := osIoctl.ioctl(c.Fd, ioctlKVMGetSRegs, uintptr(unsafe.Pointer(&sregs)))
	return sregs, err
}

func (c *VCPU) SetSRegisters(sregs SRegisters) error {
	_, err := osIoctl.ioctl(c.Fd, ioctlKVMSetSRegs, uintptr(unsafe.Pointer(&sregs)))
	return err
}
