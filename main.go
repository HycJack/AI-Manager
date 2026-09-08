package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"ai-manager/internal/app"
	"ai-manager/internal/config"
	"ai-manager/internal/logger"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg, err := config.Load(config.DefaultPath())
	if err != nil {
		log.Printf("warning: failed to load config: %v", err)
	}
	if cfg == nil {
		cfg = config.Default()
	}
	_ = cfg.EnsureDirs()

	// Initialise the dual-writer (stderr + rotating file) logger before
	// anything else so early startup messages are captured.
	if err := logger.Init(cfg.DataDir, cfg.LogLevel); err != nil {
		log.Printf("warning: failed to init logger: %v", err)
	}

	state := app.New(cfg)

	wailsApp := application.New(application.Options{
		Name:        cfg.AppName,
		Description: "AI-Manager: Agent Skills manager",
		Services: []application.Service{
			// Register each bound service here. The frontend calls their
			// exported methods through the generated bindings.
			application.NewService(app.NewGreeterService(state)),
			// Persists the frontend's preferences (theme, general settings, …)
			// as a JSON bag in the app data dir.
			application.NewService(app.NewSettingsService(state)),
			// Exposes build-time version/commit/buildtime to the frontend.
			application.NewService(app.NewVersionService()),
			// Auto-update via GitHub Releases (lazy-inits on first call).
			application.NewService(app.NewUpdateService(state)),
			// Cross-platform "launch at login" control.
			application.NewService(app.NewAutostartService(state)),
			// Recursive file search with directory / file-type filtering.
			application.NewService(app.NewSearchService(state)),
		},
		// Single instance: a second launch brings the existing window to the
		// front instead of starting a new process.
		SingleInstance: &application.SingleInstanceOptions{
			UniqueID: "ai-manager.app.single-instance",
			OnSecondInstanceLaunch: func(data application.SecondInstanceData) {
				state.ShowMainWindow()
			},
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// Keep running in the background when the window is closed so the
			// tray and global shortcuts keep working.
			ApplicationShouldTerminateAfterLastWindowClosed: false,
		},
	})

	state.SetApp(wailsApp)

	mainWindow := wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  cfg.AppName,
		Width:  1000,
		Height: 700,
		// Frameless on every platform: the in-app <header> (with the
		// app-region: drag CSS) becomes the title bar.
		Frameless: true,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 50,
			Backdrop:                application.MacBackdropTranslucent,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		Windows: application.WindowsWindow{
			// Lets WebView2 honour the CSS `-webkit-app-region: drag` on the
			// custom header via native non-client hit testing.
			NonClientRegionSupport: true,
		},
		BackgroundColour: application.NewRGB(10, 12, 22),
		URL:              "/",
	})
	state.SetMainWindow(mainWindow)

	// System tray + menu + global shortcuts.
	app.SetupTrayAndShortcuts(state)

	if err := wailsApp.Run(); err != nil {
		log.Fatal(err)
	}
}
