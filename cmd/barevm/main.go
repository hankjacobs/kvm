package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/hankjacobs/kvm"
	"github.com/hankjacobs/kvm/asmbuilder"
)

func main() {
	fBin := flag.Bool("bin", false, "Specified file is a binary. Assembler step will be skipped.")
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		fmt.Println("please specify asm file")
		os.Exit(1)
	}

	file, err := ioutil.ReadFile(args[0])
	if err != nil {
		panic(err.Error())
	}

	var bin []byte
	if *fBin {
		bin = file
	} else {
		bin, err = asmbuilder.Build(file)
		if err != nil {
			panic(err.Error())
		}
	}

	doKvm(bin)
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

	err = vm.MapUserMemory(0, 0, 0, data)
	if err != nil {
		panic(err.Error())
	}

	// step 4
	vcpu, err := vm.CreateVCPU()
	if err != nil {
		panic(err.Error())
	}

	// step 6
	regs, err := vcpu.GetRegisters()
	if err != nil {
		panic(err.Error())
	}

	regs.Rip = uint64(userEntry)
	regs.Rsp = 0x200000
	regs.Rflags = 0x2
	regs.Rax = 0x2
	regs.Rbx = 0x2
	fmt.Printf("regs %+v\n", regs)

	err = vcpu.SetRegisters(regs)
	if err != nil {
		panic(err.Error())
	}

	sregs, err := vcpu.GetSRegisters()
	if err != nil {
		panic(err.Error())
	}

	sregs.Cs.Base = 0
	sregs.Cs.Selector = 0
	fmt.Printf("sregs %+v\n", sregs)
	err = vcpu.SetSRegisters(sregs)
	if err != nil {
		panic(err.Error())
	}

	resume := make(chan struct{})
	tid := 0
	wg := sync.WaitGroup{}
	eg := sync.WaitGroup{}
	wg.Add(1)
	eg.Add(1)
	go func() {
		runtime.LockOSThread()
		tid = syscall.Gettid()
		wg.Done()

		done := false

		for !done {
			exitReason, err := vcpu.Run()
			if err != nil {
				if err == syscall.EINTR {
					fmt.Println("paused")
					<-resume
					fmt.Println("resuming")
					continue
				}

				panic(err.Error())
			}

			switch exitReason {
			case kvm.ExitReasonHlt:
				fmt.Println("Halted")
				done = true
			case kvm.ExitReasonIO:
				fmt.Printf("IO: %+v\n", vcpu.ExitIO())
			case kvm.ExitReasonFailEntry:
				panic(vcpu.ExitFailEntry().HardwareEntryFailureReason)
			case kvm.ExitReasonInternalError:
				panic(vcpu.ExitInternalError().Suberror)
			case kvm.ExitReasonShutdown:
				fmt.Println("Shutdown")
				done = true
			default:
				panic(fmt.Sprintf("unhandled %d", exitReason))
			}
		}

		eg.Done()
	}()

	wg.Wait()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	go func() {
		paused := false
		for {
			<-c
			if !paused {
				err = syscall.Kill(tid, syscall.SIGUSR2)
				if err != nil {
					panic(err.Error())
				}
				paused = true
			} else {
				resume <- struct{}{}
				paused = false
			}
		}
	}()

	eg.Wait()
	fmt.Println("DONE... YAY")
}
