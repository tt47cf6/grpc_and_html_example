package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"tt47cf6/minecraft/server"
)

var (
	frontEndPort      = flag.Int("front_end_port", 5000, "The port to run the Envot proxy on")
	frontEndAdminPort = flag.Int("front_end_admin_port", 5001, "The admin port for Envoy")
	rpcPort           = flag.Int("rpc_port", 5002, "The port to run the gRPC server on")
	htmlPort          = flag.Int("html_port", 5003, "The port to serve HTML resources on")
)

func currentDir() (string, error) {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("runtime.Caller returned not ok")
	}
	return filepath.Dir(f), nil
}

func main() {
	cwd, err := currentDir()
	if err != nil {
		log.Fatalf("currentDir(): %v", err)
	}

	rpcs := server.NewRPCServer()
	fe := server.NewFrontEnd(filepath.Join(cwd, "envoy.yaml.tmpl"))
	html := server.NewHTMLServer(filepath.Join(cwd, "web"))

	inttChan := make(chan os.Signal, 1)
	signal.Notify(inttChan, os.Interrupt)
	go func() {
		s := <-inttChan
		log.Printf("Caught signal %v, stopping", s)
		stopFunc := func(label string, ctx context.Context, f func(context.Context) error) {
			if err := f(ctx); err != nil {
				log.Printf("%s Stop(): %v", label, err)
			}

		}
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		go stopFunc("rpc", stopCtx, rpcs.Stop)
		go stopFunc("front end", stopCtx, fe.Stop)
		go stopFunc("html", stopCtx, html.Stop)

		<-stopCtx.Done()
	}()

	serveWG := &sync.WaitGroup{}

	serveWG.Add(1)
	go func() {
		defer log.Print("rpc server stopped")
		defer serveWG.Done()
		if err := rpcs.BlockingServe(*rpcPort); err != nil {
			log.Printf("rpc host BlockingServe(): %v", err)
		}
	}()

	serveWG.Add(1)
	go func() {
		defer log.Print("html server stopped")
		defer serveWG.Done()
		if err := html.BlockingServe(*htmlPort); err != nil {
			log.Printf("html BlockingServe(): %v", err)
		}
	}()

	serveWG.Add(1)
	go func() {
		defer log.Print("front end server stopped")
		defer serveWG.Done()
		if err := fe.BlockingServe(*frontEndPort, *frontEndAdminPort, *rpcPort, *htmlPort); err != nil {
			log.Printf("front end BlockingServe(): %v", err)
		}
	}()

	serveWG.Wait()
}
