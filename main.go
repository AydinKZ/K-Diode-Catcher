package main

import (
	"fmt"
	"github.com/AydinKZ/K-Diode-Catcher/config"
	"github.com/AydinKZ/K-Diode-Catcher/internal/adapters"
	"github.com/AydinKZ/K-Diode-Catcher/internal/application"
	"github.com/AydinKZ/K-Diode-Catcher/internal/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	cfg, err := config.Init("config.json")
	if err != nil {
		panic(err)
	}

	go func(cfg *config.Config) {
		hashCalculator := adapters.NewSHA1HashCalculator()

		udpReceiver, err := adapters.NewUDPReceiver(cfg.UdpAddress.Ip, cfg.UdpAddress.Port)
		if err != nil {
			panic(err)
		}

		catcherService := application.NewCatcherService(udpReceiver, hashCalculator, cfg.Queue)

		err = catcherService.ReceiveAndPublishMessages()
		if err != nil {
			panic(err)
		}
	}(cfg)

	srv, err := http.NewServer(cfg)
	if err != nil {
		panic(err)
	}

	startServerErrorCH := srv.Start()

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err = <-startServerErrorCH:
		{
			panic(err)
		}
	case q := <-quit:
		{
			fmt.Printf("receive signal %s, stopping server...\n", q.String())
			if err = srv.Stop(); err != nil {
				fmt.Printf("stop server error: %s\n", err.Error())
			}
		}
	}
}
