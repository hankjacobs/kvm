package kvm

import (
	"reflect"
	"unsafe"
)

type runData struct {
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

func (k *runData) exitUnknown() ExitUnknown {
	exit := (*ExitUnknown)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *runData) exitException() ExitException {
	exit := (*ExitException)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *runData) exitIO() ExitIO {
	exit := (*ExitIO)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *runData) exitIOData() []byte {
	io := k.exitIO()
	dataStart := unsafe.Pointer(uintptr(unsafe.Pointer(k)) + uintptr(io.DataOffset))

	var srcData []byte
	hdr := (*reflect.SliceHeader)(unsafe.Pointer(&srcData))
	hdr.Data = uintptr(dataStart)
	hdr.Len = int(io.Size)
	hdr.Cap = int(io.Size)

	data := make([]byte, len(srcData))
	copy(data, srcData)

	return data
}

func (k *runData) exitFailEntry() ExitFailEntry {
	exit := (*ExitFailEntry)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}

func (k *runData) exitInternalError() ExitInternalError {
	exit := (*ExitInternalError)(unsafe.Pointer(&k.exitReasonDataUnion[0]))
	return *exit
}
