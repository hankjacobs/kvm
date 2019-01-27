package main

import (
	"fmt"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/hankjacobs/kvm"
	"golang.org/x/sys/unix"
)

const (
	kvmGetAPIVersion          = 0xAE00
	kvmCreateVM               = 0xAE01
	kvmGetMSRIndexList        = 0xC004AE02
	kvmS390EnableSIE          = 0xAE06
	kvmCheckExtension         = 0xAE03
	kvmGetVCPUMMAPSize        = 0xAE04
	kvmGetSupportedCPUID      = 0xC008AE05
	kvmGetEmulatedCPUID       = 0xC008AE09
	kvmGetMSRFeatureIndexList = 0xC004AE0A
	kvmSetUserMemoryRegion    = 0x4020AE46
	kvmCreateVCPU             = 0xAE41
	kvmGetRegs                = 0x8090AE81
	kvmSetRegs                = 0x4090ae82
	kvmGetSregs               = 0x8138AE83
	kvmSetSregs               = 0x4138AE84
	kvmRun                    = 0xAE80
)

type kvmUserspaceMemoryRegion struct {
	Slot          uint32
	Flags         uint32
	GuestPhysAddr uint64
	MemorySize    uint64
	UserspaceAddr uint64
}

type kvmRegs struct {
	rax, rbx, rcx, rdx uint64
	rsi, rdi, rsp, rbp uint64
	r8, r9, r10, r11   uint64
	r12, r13, r14, r15 uint64
	rip, rflags        uint64
}

type kvmSegment struct {
	base                           uint64
	limit                          uint32
	selector                       uint16
	_type                          uint8
	present, dpl, db, s, l, g, avl uint8
	unusable                       uint8
	padding                        uint8
}

type kvmDtable struct {
	base    uint64
	limit   uint16
	padding [3]uint16
}

const kvmNRInterrupts = 256

type kvmSregs struct {
	cs, ds, es, fs, gs, ss  kvmSegment
	tr, ldt                 kvmSegment
	gdt, idt                kvmDtable
	cr0, cr2, cr3, cr4, cr8 uint64
	efer                    uint64
	apicBase                uint64
	interruptBitmap         [(kvmNRInterrupts + 63) / 64]uint64
}

const (
	KVMExitReasonUnknown = iota
	KVMExitReasonException
	KVMExitReasonIo
	KVMExitReasonHypercall
	KVMExitReasonDebug
	KVMExitReasonHlt
	KVMExitReasonMmio
	KVMExitReasonIrqWindowOpen
	KVMExitReasonShutdown
	KVMExitReasonFailEntry
	KVMExitReasonIntr
	KVMExitReasonSetTpr
	KVMExitReasonTprAccess
	KVMExitReasonS390Sieic
	KVMExitReasonS390Reset
	KVMExitReasonDcr
	KVMExitReasonNmi
	KVMExitReasonInternalError
	KVMExitReasonOsi
	KVMExitReasonPaprHcall
	KVMExitReasonS390Ucontrol
	KVMExitReasonWatchdog
	KVMExitReasonS390Tsch
	KVMExitReasonEpr
	KVMExitReasonSystemEvent
	KVMExitReasonS390Stsi
	KVMExitReasonIoapicEoi
	KVMExitReasonHyperv
)

type exitDirection uint8

const (
	exitDirectionIn  exitDirection = 0
	exitDirectionOut exitDirection = 1
)

type kvmExitIO struct {
	direction  exitDirection
	size       uint8
	port       uint16
	count      uint32
	dataOffset uint64
}

type kvmExitUnknown struct {
	hardwareExitReason uint64
}

type kvmExitFailEntry struct {
	hardwareEntryFailureReason uint64
}

type kvmExitException struct {
	exception uint32
	errorCode uint32
}

type kvmExitInternalError struct {
	suberror uint32
	ndata    uint32
	data     [16]uint64
}

type kvmRunData struct {
	// in
	requestInterrupWindow uint8
	immediateExit         uint8
	_                     [6]uint8

	// out
	exitReason                 uint32
	readyForInterruptInjection uint8
	ifFlag                     uint8
	flags                      uint16

	// in (pre_kvm_run), out (post_kvm_run)
	cr8      uint64
	apicBase uint64

	// 	#ifdef __KVM_S390
	// 	/* the processor status word for s390 */
	// 	__u64 psw_mask; /* psw upper half */
	// 	__u64 psw_addr; /* psw lower half */
	//  #endif

	exitReasonDataUnion [256]byte

	kvmValidRegs uint64
	kvmDirtyRegs uint64

	synRegsUnion [2048]byte
}

func (k *kvmRunData) ExitUnknown() kvmExitUnknown {
	exit := (*kvmExitUnknown)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *kvmRunData) ExitException() kvmExitException {
	exit := (*kvmExitException)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *kvmRunData) ExitIO() kvmExitIO {
	exit := (*kvmExitIO)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *kvmRunData) ExitIOData() []byte {
	io := k.ExitIO()
	dataStart := unsafe.Pointer(uintptr(unsafe.Pointer(k)) + uintptr(io.dataOffset))

	var srcData []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&srcData))
	hdr.Data = uintptr(dataStart)
	hdr.Len = int(io.size)
	hdr.Cap = int(io.size)

	data := make([]byte, len(srcData))
	copy(data, srcData)

	return data
}

func (k *kvmRunData) ExitFailEntry() kvmExitFailEntry {
	exit := (*kvmExitFailEntry)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *kvmRunData) ExitInternalError() kvmExitInternalError {
	exit := (*kvmExitInternalError)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func main() {
	doKvm([]uint8("\xB0\x61\xBA\x17\x02\xEE\xB0\n\xEE\xF4"))
}

func doKvm(code []uint8) {
	// step 1
	vm, err := kvm.CreateVM()
	if err != nil {
		panic(err.Error())
	}

	vmfd := vm.Fd
	fmt.Println("vmfd", vmfd)

	memSize := 0x40000000

	//void *mem = mmap(0, mem_size, PROT_READ|PROT_WRITE,
	//	MAP_SHARED|MAP_ANONYMOUS, -1, 0);
	// step 3
	data, err := syscall.Mmap(-1, 0, memSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED|syscall.MAP_ANONYMOUS)
	if err != nil {
		panic(err.Error())
	}
	defer syscall.Munmap(data)

	userEntry := 0x0
	copy(data[userEntry:], code)
	userspaceAddr := unsafe.Pointer(&data[0])
	region := kvmUserspaceMemoryRegion{
		Slot:          0,
		Flags:         0,
		GuestPhysAddr: 0,
		MemorySize:    uint64(memSize),
		UserspaceAddr: uint64(uintptr(userspaceAddr)),
	}

	_, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vmfd),
		uintptr(kvmSetUserMemoryRegion),
		uintptr(unsafe.Pointer(&region)),
	)
	if errno != 0 {
		panic(err.Error())
	}

	// step 4
	vcpufd, _, errno := unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vmfd),
		uintptr(kvmCreateVCPU),
		0,
	)
	if errno != 0 {
		panic(err.Error())
	}

	fmt.Println("vcpufd", vcpufd)

	// step 5
	vcpuMmapSize := vm.MMapSize
	fmt.Println("vcpuMmapSize", vcpuMmapSize)
	runMap, err := syscall.Mmap(int(vcpufd), 0, int(vcpuMmapSize), syscall.PROT_READ|syscall.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		panic(err.Error())
	}

	run := (*kvmRunData)(unsafe.Pointer(&runMap[0]))

	// step 6
	regs := kvmRegs{}
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vcpufd),
		uintptr(kvmGetRegs),
		uintptr(unsafe.Pointer(&regs)),
	)
	if errno != 0 {
		panic(err.Error())
	}
	regs.rip = uint64(userEntry)
	regs.rsp = 0x200000
	regs.rflags = 0x2
	fmt.Printf("regs %+v\n", regs)
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vcpufd),
		uintptr(kvmSetRegs),
		uintptr(unsafe.Pointer(&regs)),
	)
	if errno != 0 {
		panic(err.Error())
	}

	sregs := kvmSregs{}
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vcpufd),
		uintptr(kvmGetSregs),
		uintptr(unsafe.Pointer(&sregs)),
	)
	if errno != 0 {
		panic(err.Error())
	}
	sregs.cs.base = 0
	sregs.cs.selector = 0
	fmt.Printf("sregs %+v\n", sregs)
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vcpufd),
		uintptr(kvmSetSregs),
		uintptr(unsafe.Pointer(&sregs)),
	)
	if errno != 0 {
		panic(err.Error())
	}

	// step 7
	for {
		_, _, errno = unix.Syscall(
			unix.SYS_IOCTL,
			uintptr(vcpufd),
			uintptr(kvmRun),
			0,
		)
		if errno != 0 {
			panic(err.Error())
		}

		switch run.exitReason {
		case KVMExitReasonHlt:
			fmt.Println("Halted")
			return
		case KVMExitReasonIo:
			fmt.Println("IO:")
			fmt.Println(string(run.ExitIOData()))
		case KVMExitReasonFailEntry:
			panic(run.ExitFailEntry().hardwareEntryFailureReason)
		case KVMExitReasonInternalError:
			panic(run.ExitInternalError().suberror)
		case KVMExitReasonShutdown:
			fmt.Println("Shutdown")
		default:
			panic(fmt.Sprintf("undhandled %d", run.exitReason))
		}
		break
	}

	regs = kvmRegs{}
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vcpufd),
		uintptr(kvmGetRegs),
		uintptr(unsafe.Pointer(&regs)),
	)
	if errno != 0 {
		panic(err.Error())
	}

	fmt.Printf("regs %+v\n", regs)

	sregs = kvmSregs{}
	_, _, errno = unix.Syscall(
		unix.SYS_IOCTL,
		uintptr(vcpufd),
		uintptr(kvmGetSregs),
		uintptr(unsafe.Pointer(&sregs)),
	)
	if errno != 0 {
		panic(err.Error())
	}
	fmt.Printf("sregs %+v\n", sregs)

}
