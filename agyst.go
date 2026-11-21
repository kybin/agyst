package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"github.com/creack/pty"
	"golang.org/x/term"
)

const greenDot = "\x1b[32m●\x1b[39m"

func copyReplacingBraille(dst io.Writer, src io.Reader) {
	buf := make([]byte, 4096)
	var pending []byte
	var out bytes.Buffer

	for {
		n, err := src.Read(buf)
		if n > 0 {
			data := append(pending, buf[:n]...)
			out.Reset()
			for len(data) > 0 && utf8.FullRune(data) {
				r, size := utf8.DecodeRune(data)
				if r >= 0x2801 && r <= 0x28FF {
					out.WriteString(greenDot)
				} else {
					out.Write(data[:size])
				}
				data = data[size:]
			}
			pending = append(pending[:0], data...)
			if out.Len() > 0 {
				_, _ = dst.Write(out.Bytes())
			}
		}
		if err != nil {
			break
		}
	}
	if len(pending) > 0 {
		_, _ = dst.Write(pending)
	}
}

func run() int {
	stdinFD, stdoutFD := int(os.Stdin.Fd()), int(os.Stdout.Fd())

	agyPath, err := exec.LookPath("agy")
	if err != nil {
		fmt.Fprintf(os.Stderr, "agyst: agy not found in PATH: %v\n", err)
		return 127
	}
	if !term.IsTerminal(stdinFD) || !term.IsTerminal(stdoutFD) {
		if err := syscall.Exec(agyPath, append([]string{"agy"}, os.Args[1:]...), os.Environ()); err != nil {
			fmt.Fprintf(os.Stderr, "agyst: exec agy: %v\n", err)
			return 1
		}
	}

	cmd := exec.Command(agyPath, os.Args[1:]...)
	var ptmx *os.File
	if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
		ptmx, err = pty.StartWithSize(cmd, ws)
	} else {
		ptmx, err = pty.Start(cmd)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "agyst: start agy: %v\n", err)
		return 1
	}
	defer ptmx.Close()

	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	defer signal.Stop(winch)
	go func() {
		for range winch {
			_ = pty.InheritSize(os.Stdin, ptmx)
		}
	}()

	oldState, err := term.MakeRaw(stdinFD)
	if err != nil {
		fmt.Fprintf(os.Stderr, "agyst: raw mode: %v\n", err)
		return 1
	}
	defer term.Restore(stdinFD, oldState)

	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	copyReplacingBraille(os.Stdout, ptmx)

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

func main() {
	os.Exit(run())
}
