package asmbuilder

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// Build builds the asm and outputs a byte array suitable for being executed by kvm
func Build(asm []byte) ([]byte, error) {
	td, err := ioutil.TempDir("", "asmbuilder")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(td)

	opath := filepath.Join(td, "asm.o")
	cmd := exec.Command("as", "-32", "-o", opath, "--")
	cmd.Stdin = bytes.NewReader(asm)
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	binary := bytes.Buffer{}
	cmd = exec.Command("ld", "-m", "elf_i386", "--oformat=binary", "-e", "_start", "-o", "/dev/stdout", opath)
	cmd.Stdout = &binary
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return binary.Bytes(), nil
}
