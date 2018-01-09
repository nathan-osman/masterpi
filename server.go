package masterpi

import (
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	relay     *Relay
	listener  net.Listener
	log       *logrus.Entry
	stoppedCh chan bool
}

func (s *Server) apiLampToggle(w http.ResponseWriter, r *http.Request) {
	s.relay.Toggle()
	http.Redirect(w, r, "/", http.StatusFound)
}

func NewServer(r *Relay) (*Server, error) {
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		return nil, err
	}
	var (
		router = mux.NewRouter()
		s      = &Server{
			relay:     r,
			listener:  l,
			log:       logrus.WithField("context", "server"),
			stoppedCh: make(chan bool),
		}
		server = http.Server{
			Handler: router,
		}
		handler = http.FileServer(HTTP)
	)
	router.HandleFunc("/api/lamp/toggle", s.apiLampToggle)
	router.PathPrefix("/").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.URL.Path = "/www" + r.URL.Path
			handler.ServeHTTP(w, r)
		},
	)
	go func() {
		defer close(s.stoppedCh)
		defer s.log.Info("server shut down")
		s.log.Info("server started")
		if err := server.Serve(l); err != nil {
			s.log.Error(err.Error())
		}
	}()
	return s, nil
}

func (s *Server) Close() {
	s.listener.Close()
	<-s.stoppedCh
}
