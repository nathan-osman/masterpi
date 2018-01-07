package masterpi

import (
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

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	s.writeResponse(w, "text/html", []byte(`<!DOCTYPE html>
<html>
    <head>
        <meta charset="utf-8">
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
    </head>
    <body>
        <form method="post" action="/api/lamp/toggle">
            <button type="submit">Toggle Lamp</button>
        </form>
    </body>
</html>
`))
}

func (s *Server) apiLampToggle(w http.ResponseWriter, r *http.Request) {
	s.relay.Toggle()
	http.Redirect(w, r, "/", http.StatusFound)
}

func NewServer(r *Relay) (*Server, error) {
	l, err := net.Listen("tcp", ":80")
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
	)
	router.HandleFunc("/", s.index)
	router.HandleFunc("/api/lamp/toggle", s.apiLampToggle)
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
