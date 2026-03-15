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
    :root{color-scheme:light;font-family:"Segoe UI",Arial,sans-serif}
    *{box-sizing:border-box}
    html,body{margin:0;width:100%;height:100%;overflow:hidden;background:#f2f2f2;color:#171717}
    body{display:grid;place-items:center}
    .wizard{width:min(920px,calc(100vw - 32px));height:min(640px,calc(100vh - 32px));display:grid;grid-template-columns:240px minmax(0,1fr);background:#ffffff;border:1px solid #d4d4d4;border-radius:18px;box-shadow:0 24px 60px rgba(0,0,0,.08);overflow:hidden}
    .sidebar{background:#f7f7f7;border-right:1px solid #e5e5e5;padding:28px 22px;display:grid;grid-template-rows:auto 1fr auto;gap:24px}
    .brand p,.step-label,.muted{margin:0;color:#737373;font-size:12px;letter-spacing:.12em;text-transform:uppercase}
    .brand h1,.content h2,.panel h3{margin:8px 0 0;font-size:28px;line-height:1.1}
    .steps{display:grid;gap:12px;align-content:start}
    .step{padding:12px 14px;border-radius:12px;border:1px solid transparent;background:transparent;font-size:14px}
    .step.active{background:#ffffff;border-color:#d4d4d4;font-weight:600}
    .content{padding:28px;display:grid;grid-template-rows:auto auto auto 1fr auto;gap:18px}
    .lead{max-width:600px;margin:0;color:#525252;line-height:1.5}
    .options{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px}
    .option{padding:18px;border:1px solid #d4d4d4;border-radius:14px;background:#fafafa;text-align:left}
    .option strong{display:block;font-size:16px;margin-bottom:6px}
    .option p{margin:0;color:#666;line-height:1.4}
    .option.active{background:#171717;color:#fff;border-color:#171717}
    .option.active p{color:#d4d4d4}
    .panel{border:1px solid #e5e5e5;border-radius:16px;background:#fafafa;padding:18px;display:grid;gap:12px;min-height:0}
    .progress-list{display:grid;gap:10px}
    .progress-item{display:flex;justify-content:space-between;gap:12px;font-size:14px;color:#525252}
    .progress-item.active{color:#171717;font-weight:600}
    .progress-item.done{color:#525252}
    .output{min-height:84px;max-height:120px;padding:14px;border-radius:12px;background:#f1f1f1;border:1px solid #e5e5e5;font-family:Consolas,monospace;font-size:13px;overflow:hidden;white-space:pre-wrap}
    .footer{display:flex;justify-content:space-between;align-items:center;gap:16px}
    .status{color:#525252;margin:0}
    .actions{display:flex;gap:12px}
    button{padding:12px 18px;border-radius:999px;border:1px solid #d4d4d4;background:#fff;color:#171717;font-weight:600;cursor:pointer}
    button.primary{background:#171717;color:#fff;border-color:#171717}
    button:disabled{opacity:.5;cursor:not-allowed}
  </style>
</head>
<body>
  <div class="wizard">
    <aside class="sidebar">
      <div class="brand">
        <p>AWaN Installer</p>
        <h1>Setup</h1>
      </div>
      <div class="steps">
        <div class="step active">1. Welcome</div>
        <div class="step active">2. Choose interface</div>
        <div class="step">3. Install components</div>
        <div class="step">4. Finish</div>
      </div>
      <div>
        <p class="step-label">Target path</p>
        <p id="root-path" class="status">Loading...</p>
      </div>
    </aside>
    <main class="content">
      <div>
        <p class="muted">Standard installer</p>
        <h2>Install AWaN</h2>
      </div>
      <p class="lead">Install AWaN Core and choose the interface you want to use. This installer keeps the flow simple, fixed, and non-scrollable.</p>
      <div class="options">
        <button id="install-gui" class="option active">
          <strong>Desktop GUI</strong>
          <p>Native desktop interface for chatting, viewing memory, and managing files.</p>
        </button>
        <button id="install-tui" class="option">
          <strong>Terminal TUI</strong>
          <p>Lightweight terminal client for servers, SSH sessions, and keyboard-first use.</p>
        </button>
      </div>
      <section class="panel">
        <h3>Installation progress</h3>
        <div id="progress-list" class="progress-list">
          <div class="progress-item"><span>Download AWaN Core</span><span>Pending</span></div>
          <div class="progress-item"><span>Download selected interface</span><span>Pending</span></div>
          <div class="progress-item"><span>Configure system</span><span>Pending</span></div>
          <div class="progress-item"><span>Create launcher</span><span>Pending</span></div>
        </div>
        <div id="output" class="output">Ready to install.</div>
      </section>
      <div class="footer">
        <p id="finish" class="status">Choose an interface and start the installation.</p>
        <div class="actions">
          <button id="cancel">Close</button>
          <button id="start-install" class="primary">Install</button>
        </div>
      </div>
    </main>
  </div>
  <script>
    const output = document.getElementById('output');
    const finish = document.getElementById('finish');
    const rootPath = document.getElementById('root-path');
    const progressList = document.getElementById('progress-list');
    const startButton = document.getElementById('start-install');
    const guiButton = document.getElementById('install-gui');
    const tuiButton = document.getElementById('install-tui');
    const cancelButton = document.getElementById('cancel');
    let selectedTarget = 'gui';

    function selectTarget(target) {
      selectedTarget = target;
      guiButton.classList.toggle('active', target === 'gui');
      tuiButton.classList.toggle('active', target === 'tui');
    }

    function setProgress(state) {
      const statuses = {
        idle: ['Pending', 'Pending', 'Pending', 'Pending'],
        running: ['Running', 'Waiting', 'Waiting', 'Waiting'],
        done: ['Done', 'Done', 'Done', 'Done']
      };
      const values = statuses[state] || statuses.idle;
      progressList.innerHTML = [
        'Download AWaN Core',
        'Download selected interface',
        'Configure system',
        'Create launcher'
      ].map((label, index) => '<div class="progress-item ' + (state === 'done' ? 'done' : index === 0 && state === 'running' ? 'active' : '') + '"><span>' + label + '</span><span>' + values[index] + '</span></div>').join('');
    }

    async function bootstrap() {
      try {
        const info = await window.go.gui.App.SystemInfo();
        rootPath.textContent = info.rootPath;
      } catch (error) {
        rootPath.textContent = 'Unavailable';
      }
    }

    async function runInstall(){
      output.textContent = 'Installing ' + selectedTarget + '...';
      finish.textContent = 'Installing components...';
      startButton.disabled = true;
      setProgress('running');
      try {
        const result = await window.go.gui.App.Install(selectedTarget);
        output.textContent = JSON.stringify(result, null, 2);
        finish.textContent = 'Installation complete for ' + result.target + ' at ' + result.rootPath;
        setProgress('done');
      } catch (error) {
        output.textContent = String(error);
        finish.textContent = 'Installation failed.';
        setProgress('idle');
      } finally {
        startButton.disabled = false;
      }
    }
    guiButton.addEventListener('click', () => selectTarget('gui'));
    tuiButton.addEventListener('click', () => selectTarget('tui'));
    startButton.addEventListener('click', runInstall);
    cancelButton.addEventListener('click', () => window.close());
    bootstrap();
  </script>
</body>
</html>`

	return fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: []byte(html),
		},
	}
}
