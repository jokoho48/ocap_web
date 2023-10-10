package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/OCAP2/web/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/crypto/acme/autocert"
)

func main() {
	if err := app(); err != nil {
		log.Panicln(err)
	}
}

func app() error {
	setting, err := server.NewSetting()
	if err != nil {
		return fmt.Errorf("setting: %w", err)
	}

	operation, err := server.NewRepoOperation(setting.DB)
	if err != nil {
		return fmt.Errorf("operation: %w", err)
	}

	marker, err := server.NewRepoMarker(setting.Markers)
	if err != nil {
		return fmt.Errorf("marker: %w", err)
	}

	ammo, err := server.NewRepoAmmo(setting.Ammo)
	if err != nil {
		return fmt.Errorf("ammo: %w", err)
	}

	e := echo.New()

	loggerConfig := middleware.DefaultLoggerConfig
	if setting.Logger {
		flog, err := os.OpenFile("ocap.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("open logger file: %w", err)
		}
		defer flog.Close()

		loggerConfig.Output = io.MultiWriter(os.Stdout, flog)
	}

	e.Use(
		middleware.LoggerWithConfig(loggerConfig),
	)

	server.NewHandler(e, operation, marker, ammo, setting)
	if setting.UseHttps {
		if setting.UseAutoTLS {
			autoTLSManager := autocert.Manager{
				Prompt: autocert.AcceptTOS,
				// Cache certificates to avoid issues with rate limits (https://letsencrypt.org/docs/rate-limits)
				Cache: autocert.DirCache("/var/www/.cache"),
			}
			e.AutoTLSManager = autoTLSManager
			err = e.StartAutoTLS(setting.Listen)
			if err != nil {
				return fmt.Errorf("start server: %w", err)
			}
			return nil
		}

		err = e.StartTLS(setting.Listen, setting.CertFile, setting.KeyFile)
		if err != nil {
			return fmt.Errorf("start server: %w", err)
		}
		return nil
	}
	err = e.Start(setting.Listen)
	if err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return nil
}
