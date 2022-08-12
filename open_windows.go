//go:build windows

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func open(port int) {
	log.Println("windows打开浏览器:", exec.Command("explorer", fmt.Sprintf("http://localhost:%d", port)).Run())
}
