// Copyright (c) 2020 Byron Grobe
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/go-pipe/pipe"
	"github.com/mitchellh/go-homedir"
	"github.com/sean-/seed"
)

func main() {
	var (
		baseImage *string
		name      *string
		kill      *bool
		args      []string
		home      string
		driveFile string
		e         error
		vmsString string
		vmsBytes  []byte
		vms       []string
		pipeline  pipe.Pipe
		strbld    strings.Builder
		cmd       *exec.Cmd
	)

	seed.MustInit()

	baseImage = flag.String("base", "openbsd66_amd64-base", "Base image for VM")
	name = flag.String("name", fmt.Sprintf("testvm-%d", rand.Uint32()), "Name for VM")
	kill = flag.Bool("kill", false, "Kill VM")
	flag.Parse()

	if *kill {
		pipeline = pipe.Line(
			pipe.Exec("/usr/sbin/vmctl", "status"),
			pipe.Exec("/usr/bin/sed", "-E", "-e", "s/[[:space:][:blank:]]{2,}/ /g;/NAME$/d"),
			pipe.Exec("/usr/bin/cut", "-d", " ", "-f10"),
			pipe.Exec("/usr/bin/sed", "-n", "-e", "/testvm/p;s/\\.[[:alnum:]]*$//"),
		)

		vmsBytes, _ = pipe.Output(pipeline)

		_, e = strbld.Write(vmsBytes)
		if e != nil {
			log.Fatal(e)
		}

		vmsString = strbld.String()

		vms = strings.Split(vmsString, "\n")

		for _, vm := range vms {
			args = []string{
				"/usr/bin/vmctl",
				"stop", "-f", vm,
			}

			cmd = exec.Command("/usr/sbin/vmctl")
			cmd.Args = args
			_ = cmd.Run()
		}

		return
	}

	home, e = homedir.Dir()
	if e != nil {
		log.Fatal(e)
	}

	driveFile = path.Join(home, ".cache", "testvm", *name+".qcow2")

	args = []string{
		"/usr/sbin/vmctl",
		"create", "-b", path.Join("/usr/local/vm", *baseImage + ".qcow2"),
		driveFile,
	}

	log.Println(args)

	cmd = exec.Command("/usr/sbin/vmctl")
	cmd.Args = args
	e = cmd.Run()
	if e != nil {
		log.Fatal(e)
	}

	args = []string{
		"/usr/sbin/vmctl",
		"start", "-c", "-L", "-d", driveFile,
	      }

	args = append(args, flag.Args()...)
	args = append(args, *name)

	log.Println(args)

	cmd = exec.Command("/usr/sbin/vmctl")
	cmd.Args = args
	e = cmd.Run()
	if e != nil {
		log.Fatal(e)
	}

	args = []string{
		"/usr/sbin/vmctl",
		"console", *name,
	}

	cmd = exec.Command("/usr/sbin/vmctl")
	cmd.Args = args
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	e = cmd.Run()
	if e != nil {
		log.Fatal(e)
	}

	e = os.Remove(driveFile)
	if e != nil {
		log.Fatal(e)
	}

}
