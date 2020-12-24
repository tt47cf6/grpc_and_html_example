package server

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

var (
	copyFEOut = flag.Bool("copy_frontend_out", false, "Set to copy Envoy's stdout and stderr to this'")
)

type FrontEnd struct {
	cmd   *exec.Cmd
	cmdMu sync.Mutex

	cfgFile     string
	cfgTmplPath string
}

func NewFrontEnd(cfgPath string) *FrontEnd {
	return &FrontEnd{
		cfgTmplPath: cfgPath,
	}
}

func (fe *FrontEnd) BlockingServe(fePort, feAdminPort, rpcPort, htmlPort int) error {
	replacements := map[string]string{
		"FRONT_END_PORT":       strconv.Itoa(fePort),
		"FRONT_END_ADMIN_PORT": strconv.Itoa(feAdminPort),
		"RPC_PORT":             strconv.Itoa(rpcPort),
		"HTML_PORT":            strconv.Itoa(htmlPort),
	}

	cfgFileBytes, err := ioutil.ReadFile(fe.cfgTmplPath)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile(%q): %v", fe.cfgTmplPath, err)
	}

	cfgFileStr := string(cfgFileBytes)
	for k, v := range replacements {
		cfgFileStr = strings.ReplaceAll(cfgFileStr, k, v)
	}

	f, err := ioutil.TempFile("", "envoy-*.yaml")
	if err != nil {
		return fmt.Errorf("ioutil.TempFile(): %v", err)
	}
	f.WriteString(cfgFileStr)
	f.Close()
	fe.cfgFile = f.Name()

	fe.cmdMu.Lock()
	fe.cmd = exec.Command("getenvoy", "run", "standard:1.16.2", "--", "-c", fe.cfgFile)
	fe.cmdMu.Unlock()

	if *copyFEOut {
		fe.cmd.Stdout = os.Stdout
		fe.cmd.Stderr = os.Stderr
	}

	log.Printf("starting front end server on port %d", fePort)
	return fe.cmd.Run()
}

func (fe *FrontEnd) Stop(ctx context.Context) error {
	fe.cmdMu.Lock()
	defer fe.cmdMu.Unlock()
	if fe.cmd == nil || fe.cmd.Process == nil {
		return nil
	}

	defer func() {
		if err := os.Remove(fe.cfgFile); err != nil {
			log.Printf("error: os.Remove(%q): %v", fe.cfgFile, err)
		}

	}()

	errChan := make(chan error)
	go func() {
		if err := fe.cmd.Process.Signal(os.Interrupt); err != nil {
			errChan <- fmt.Errorf("envoy.Signal(2): %v", err)
		}
		errChan <- nil
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Print("force-stopping the front end server")
		return fe.cmd.Process.Kill()
	}

	return nil
}
