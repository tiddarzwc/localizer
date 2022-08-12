//go:build darwin

package main

import (
	"fmt"
	"log"
	"os/exec"
)

func open(port int) {
	log.Println("darwin打开浏览器:", exec.Command("open", fmt.Sprintf("http://localhost:%d", port)).Run())
}
