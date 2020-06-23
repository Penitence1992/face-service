package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type GinRegisterFunc = func(engine *gin.Engine)

type Server struct {
	Port        uint16
	BindAddress string
	HttpServer  *http.Server
	route       *gin.Engine
	sslMode     bool
}

func CreateServer(port uint16, bindAddress string, tlsConfig *TlsConfig) *Server {
	address := fmt.Sprintf("%s:%d", bindAddress, port)
	httpServerConfig := &Server{
		Port:        port,
		BindAddress: bindAddress,
		HttpServer: &http.Server{
			Addr:    address,
			Handler: nil,
		},
		route: gin.Default(),
	}
	if tlsConfig != nil {
		if tlsCertKey, err := tls.LoadX509KeyPair(tlsConfig.TlsCertPath, tlsConfig.TlsKeyPath); err != nil {
			log.Fatal("loading tls error : ", err)
		} else {
			httpServerConfig.HttpServer.TLSConfig = &tls.Config{Certificates: []tls.Certificate{tlsCertKey}}
			httpServerConfig.sslMode = true
		}
	}
	return httpServerConfig
}

func (s Server) RegisterRoute(f GinRegisterFunc) {
	f(s.route)
}

func (s Server) StartListen() {
	done := make(chan os.Signal)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sg := <-done
		log.Info("accept signal {}", sg)
		if err := s.HttpServer.Shutdown(context.Background()); err != nil {
			log.Error("Shutdown server:", err)
		} else {
			log.Info("Shutdown server")
		}
	}()

	log.Info("Start http server ...")
	s.HttpServer.Handler = s.route
	var err error
	if s.sslMode {
		err = s.HttpServer.ListenAndServeTLS("", "")
	} else {
		err = s.HttpServer.ListenAndServe()
	}
	if err != nil {
		if err == http.ErrServerClosed {
			log.Warn("Server closed under request")
		} else {
			log.Info("Server closed unexpected")
		}
	}
}
