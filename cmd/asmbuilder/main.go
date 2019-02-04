package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hankjacobs/kvm/asmbuilder"
)

func main() {
	flag.Parse()

	args := flag.Args()

	fail := func(msg string) {
		fmt.Println("error: ", msg)
		os.Exit(1)
	}

	if len(args) != 2 {
		fail("please set a input file and an output file")
	}

	asm, err := ioutil.ReadFile(args[0])
	if err != nil {
		fail(fmt.Sprintf("failed reading file %s: %v", args[0], err))
	}

	bin, err := asmbuilder.Build(asm)
	if err != nil {
		fail(fmt.Sprintf("failed building asm: %v", err))
	}

	err = ioutil.WriteFile(args[1], bin, 755)
	if err != nil {
		fail(fmt.Sprintf("failed writing file %s: %v", args[1], err))
	}
}
