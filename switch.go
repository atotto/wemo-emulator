package wemo

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type SwitchService struct {
	state bool // false=off, true=on
	mu    sync.Mutex

	name        string
	port        string
	uuid        string
	serial      string
	onCallback  func(ctx context.Context, state bool) bool
	offCallback func(ctx context.Context, state bool) bool
}

func ConfigSwitchService(name string, port, uuid, serial string, onCallback, offCallback func(ctx context.Context, state bool) bool) *SwitchService {
	return &SwitchService{
		state:       false,
		name:        name,
		port:        port,
		uuid:        uuid,
		serial:      serial,
		onCallback:  onCallback,
		offCallback: offCallback,
	}
}

func (s *SwitchService) run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to wemo emulator.\nhttps://github.com/atotto/wemo-emulator")
	})
	mux.HandleFunc("/setup.xml", s.handleSetup)
	mux.HandleFunc("/upnp/control/basicevent1", s.handleUpnpControlBasicEvent1)
	mux.HandleFunc("/eventservice.xml", handleEventService)

	// TODO: shutdown

	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", s.port), mux)
}

func (s *SwitchService) SetState(state bool) {
	s.mu.Lock()
	s.state = state
	s.mu.Unlock()
}

func StartSwitchServices(ctx context.Context, services ...*SwitchService) error {
	for _, s := range services {
		s := s
		go func() {
			err := s.run(ctx)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	return startUPnPService(ctx, services...)
}
