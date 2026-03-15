package gui

import (
	"context"
	"testing/fstest"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"

	"github.com/whitehai11/AWaN-Installer/installer"
)

type App struct {
	ctx       context.Context
	installer *installer.Installer
}

func NewApp(flow *installer.Installer) *App {
	return &App{installer: flow}
}

func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SystemInfo() map[string]string {
	platform := a.installer.Platform()
	return map[string]string{
		"os":       platform.OS,
		"arch":     platform.Architecture,
		"rootPath": a.installer.Paths().Root,
	}
}

func (a *App) Install(target string) (*installer.Result, error) {
	return a.installer.Run(target, func(step installer.Step, index, total int) {})
}

// Run launches the graphical installer.
func Run(flow *installer.Installer) error {
	app := NewApp(flow)
	assets := inMemoryAssets()

	return wails.Run(&options.App{
		Title:  "AWaN Installer",
		Width:  980,
		Height: 720,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		Windows: &windows.Options{
			DisableWindowIcon: false,
		},
		Bind: []interface{}{app},
	})
}

func inMemoryAssets() fstest.MapFS {
	html := `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>AWaN Installer</title>
  <style>
    body{margin:0;font-family:Segoe UI,sans-serif;background:#f4efe5;color:#1f2a21}
    .shell{max-width:980px;margin:0 auto;padding:32px;display:grid;gap:20px}
    .panel{background:#fffaf0;border:1px solid rgba(30,60,40,.12);border-radius:20px;padding:24px;box-shadow:0 24px 80px rgba(30,60,40,.08)}
    .options{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:16px}
    button{padding:14px 18px;border:0;border-radius:999px;background:#1e5f43;color:#fff;font-weight:700;cursor:pointer}
    .secondary{background:#d8ecdf;color:#1e5f43}
    .muted{color:#5f6e62}
    pre{background:#f6f1e7;border-radius:16px;padding:16px;overflow:auto}
  </style>
</head>
<body>
  <div class="shell">
    <section class="panel">
      <p class="muted">Welcome to AWaN Installer</p>
      <h1>Select installation options</h1>
      <p>This graphical installer can install AWaN Core plus either the desktop GUI or the terminal TUI.</p>
    </section>
    <section class="panel">
      <h2>Installation options</h2>
      <div class="options">
        <button id="install-gui">Install GUI version</button>
        <button id="install-tui" class="secondary">Install TUI version</button>
      </div>
    </section>
    <section class="panel">
      <h2>Progress</h2>
      <pre id="output">Ready.</pre>
    </section>
    <section class="panel">
      <h2>Finish</h2>
      <p id="finish" class="muted">Run an installation to finish setup.</p>
    </section>
  </div>
  <script>
    const output = document.getElementById('output');
    const finish = document.getElementById('finish');
    async function runInstall(target){
      output.textContent = 'Installing ' + target + '...';
      finish.textContent = 'Running installation...';
      try {
        const result = await window.go.gui.App.Install(target);
        output.textContent = JSON.stringify(result, null, 2);
        finish.textContent = 'Installation complete for ' + result.target + ' at ' + result.rootPath;
      } catch (error) {
        output.textContent = String(error);
        finish.textContent = 'Installation failed.';
      }
    }
    document.getElementById('install-gui').addEventListener('click', () => runInstall('gui'));
    document.getElementById('install-tui').addEventListener('click', () => runInstall('tui'));
  </script>
</body>
</html>`

	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(html),
		},
	}
}
