package kvm

type ExitReason int

const (
	ExitReasonUnknown ExitReason = iota
	ExitReasonException
	ExitReasonIO
	ExitReasonHypercall
	ExitReasonDebug
	ExitReasonHlt
	ExitReasonMmio
	ExitReasonIrqWindowOpen
	ExitReasonShutdown
	ExitReasonFailEntry
	ExitReasonIntr
	ExitReasonSetTpr
	ExitReasonTprAccess
	ExitReasonS390Sieic
	ExitReasonS390Reset
	ExitReasonDcr
	ExitReasonNmi
	ExitReasonInternalError
	ExitReasonOsi
	ExitReasonPaprHcall
	ExitReasonS390Ucontrol
	ExitReasonWatchdog
	ExitReasonS390Tsch
	ExitReasonEpr
	ExitReasonSystemEvent
	ExitReasonS390Stsi
	ExitReasonIoapicEoi
	ExitReasonHyperv
)

type IODirection uint8

const (
	IODirectionIn  IODirection = 0
	IODirectionOut IODirection = 1
)

type ExitIO struct {
	Direction  IODirection
	Size       uint8
	Port       uint16
	Count      uint32
	DataOffset uint64
}

type ExitUnknown struct {
	HardwareExitReason uint64
}

type ExitFailEntry struct {
	HardwareEntryFailureReason uint64
}

type ExitException struct {
	Exception uint32
	ErrorCode uint32
}

type ExitInternalError struct {
	Suberror uint32
	Ndata    uint32
	Data     [16]uint64
}

func (c *VCPU) Run() (ExitReason, error) {
	_, err := osIoctl.ioctl(c.Fd, ioctlKVMRun, 0)
	return ExitReason(c.run.exitReason), err
}

func (c *VCPU) ExitUnknown() ExitUnknown {
	return c.run.exitUnknown()
}

func (c *VCPU) ExitException() ExitException {
	return c.run.exitException()
}

func (c *VCPU) ExitIO() ExitIO {
	return c.run.exitIO()
}

func (c *VCPU) ExitIOData() []byte {
	return c.run.exitIOData()
}

func (c *VCPU) ExitFailEntry() ExitFailEntry {
	return c.run.exitFailEntry()
}

func (c *VCPU) ExitInternalError() ExitInternalError {
	return c.run.exitInternalError()
}
