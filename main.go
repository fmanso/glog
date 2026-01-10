package main

import (
	"embed"
	"glog/db"
	"os"
	"strings"

	"github.com/labstack/gommon/log"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Configure log level from environment variable (default: INFO)
	logLevel := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch logLevel {
	case "DEBUG":
		log.SetLevel(log.DEBUG)
	case "INFO":
		log.SetLevel(log.INFO)
	case "WARN":
		log.SetLevel(log.WARN)
	case "ERROR":
		log.SetLevel(log.ERROR)
	default:
		log.SetLevel(log.INFO) // Production default
	}

	dbStore, err := db.NewDocumentStore("glog.db")
	if err != nil {
		panic("Failed to initialize database: " + err.Error())
	}

	defer dbStore.Close()

	app := NewApp(dbStore)

	// Create application with options
	err = wails.Run(&options.App{
		Title:     "glog",
		Width:     1024,
		Height:    768,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
