package app

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// SetupTrayAndShortcuts creates a system tray with a menu, registers global
// shortcuts, and makes the main window hide-to-tray instead of quitting when
// closed (so background work can keep running).
func SetupTrayAndShortcuts(state *State) {
	app := state.app

	name := "Skeleton"
	if state.cfg != nil && state.cfg.AppName != "" {
		name = state.cfg.AppName
	}

	tray := app.SystemTray.New()
	tray.SetTooltip(name)
	tray.SetIcon(generateTrayIcon())

	menu := app.NewMenu()
	menu.Add("显示主窗口").OnClick(func(ctx *application.Context) {
		state.ShowMainWindow()
	})
	menu.AddSeparator()
	menu.Add("退出").OnClick(func(ctx *application.Context) {
		app.Quit()
	})
	tray.SetMenu(menu)

	// Global shortcuts (best effort; registration failure is non-fatal).
	register := func(accel string, fn func()) {
		if err := app.GlobalShortcut.Register(accel, fn); err != nil {
			log.Printf("shortcut %s: %v", accel, err)
		}
	}
	register("CmdOrCtrl+Alt+M", func() { state.ShowMainWindow() })

	// Hide to tray instead of quitting when the main window is closed.
	if win := state.MainWindow(); win != nil {
		win.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
			win.Hide()
			e.Cancel()
		})
	}
}

// generateTrayIcon draws a small gradient rounded square usable as a tray icon.
func generateTrayIcon() []byte {
	const size = 22
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	// Transparent background.
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetRGBA(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	radius := 5
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !inRoundedRect(x, y, size, radius) {
				continue
			}
			// Vertical gradient: accent blue -> purple.
			t := float64(y) / float64(size-1)
			r := lerp(91, 124, t)
			g := lerp(140, 91, t)
			b := lerp(255, 255, t)
			img.SetRGBA(x, y, color.RGBA{uint8(r), uint8(g), uint8(b), 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		log.Printf("icon: %v", err)
		return nil
	}
	return buf.Bytes()
}

func inRoundedRect(x, y, size, radius int) bool {
	r := float64(radius)
	s := float64(size - 1)
	if x < radius && y < radius {
		return dist(x, y, r, r) <= r
	}
	if x >= size-radius && y < radius {
		return dist(x, y, s-r, r) <= r
	}
	if x < radius && y >= size-radius {
		return dist(x, y, r, s-r) <= r
	}
	if x >= size-radius && y >= size-radius {
		return dist(x, y, s-r, s-r) <= r
	}
	return true
}

func dist(x, y int, cx, cy float64) float64 {
	dx := float64(x) - cx
	dy := float64(y) - cy
	return math.Sqrt(dx*dx + dy*dy)
}

func lerp(a, b, t float64) float64 { return a + (b-a)*t }
