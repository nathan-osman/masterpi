package masterpi

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type Server struct {
	relay     *Relay
	listener  net.Listener
	log       *logrus.Entry
	stoppedCh chan bool
}

func (s *Server) writeResponse(w http.ResponseWriter, contentType string, content []byte) {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

type apiLampStateParams struct {
	Value bool `json:"value"`
}

func (s *Server) apiLampState(w http.ResponseWriter, r *http.Request) {
	err := func() error {
		switch r.Method {
		case http.MethodGet:
			b, err := json.Marshal(apiLampStateParams{Value: s.relay.IsOn()})
			if err != nil {
				return err
			}
			s.writeResponse(w, "application/json", b)
		case http.MethodPost:
			var v apiLampStateParams
			if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
				return err
			}
			s.relay.SetOn(v.Value)
			s.writeResponse(w, "application/json", []byte("{}"))
		default:
			return errors.New("invalid method")
		}
		return nil
	}()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
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
	router.HandleFunc("/api/lamp/state", s.apiLampState)
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
