package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type HTMLServer struct {
	path string

	s   *http.Server
	smu sync.Mutex
}

func NewHTMLServer(path string) *HTMLServer {
	return &HTMLServer{
		path: path,
	}
}

func (hs *HTMLServer) BlockingServe(port int) error {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(hs.path)))

	mux.HandleFunc("/MyRPCServer/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Println("got html request")
		rw.WriteHeader(404)
	})

	hs.smu.Lock()
	hs.s = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	hs.smu.Unlock()

	log.Printf("starting html server of %q on port %d", hs.path, port)
	err := hs.s.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (hs *HTMLServer) Stop(ctx context.Context) error {
	hs.smu.Lock()
	defer hs.smu.Unlock()
	if hs.s == nil {
		return nil
	}

	if err := hs.s.Shutdown(ctx); err != nil {
		return fmt.Errorf("http.Shutdown(): %v", err)
	}
	return nil
}
