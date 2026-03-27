package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/System9-Software/sysix/internal/analyzer"
	"github.com/System9-Software/sysix/internal/collector"
)

type HistoryPoint struct {
	Time        int64   `json:"time"`
	CPUPercent  float64 `json:"cpu"`
	MemPercent  float64 `json:"mem"`
	DiskPercent float64 `json:"disk"`
}

var (
	history []HistoryPoint
	histMu  sync.Mutex
)

func recordHistory() {
	for {
		snap, err := collector.GetSnapshot()
		if err == nil {
			histMu.Lock()
			history = append(history, HistoryPoint{
				Time:        time.Now().Unix(),
				CPUPercent:  snap.CPUPercent,
				MemPercent:  snap.MemPercent,
				DiskPercent: snap.DiskPercent,
			})
			if len(history) > 60 {
				history = history[len(history)-60:]
			}
			histMu.Unlock()
		}
		time.Sleep(2 * time.Second)
	}
}

func Start(port int) error {
	go recordHistory()
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/snapshot", handleSnapshot)
	http.HandleFunc("/api/ports", handlePorts)
	http.HandleFunc("/api/network", handleNetwork)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/analysis", handleAnalysis)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("sysix web UI running at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboard))
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	histMu.Lock()
	defer histMu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func handleAnalysis(w http.ResponseWriter, r *http.Request) {
	histMu.Lock()
	hist := make([]analyzer.HistoryPoint, len(history))
	for i, h := range history {
		hist[i] = analyzer.HistoryPoint{
			CPUPercent:  h.CPUPercent,
			MemPercent:  h.MemPercent,
			DiskPercent: h.DiskPercent,
		}
	}
	histMu.Unlock()

	report := analyzer.Analyze(hist)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func handleSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshot, err := collector.GetSnapshot()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(snapshot)
}

func handlePorts(w http.ResponseWriter, r *http.Request) {
	ports, err := collector.GetPorts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ports)
}

func handleNetwork(w http.ResponseWriter, r *http.Request) {
	network, err := collector.GetNetwork()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(network)
}

const dashboard = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>sysix — System9</title>
<style>
  @import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600;700&family=DM+Sans:wght@300;400;500&display=swap');

  :root {
    --bg: #090D15;
    --sidebar: #111827;
    --card: #1E2D45;
    --accent: #4DA8FF;
    --text: #E8F0FF;
    --muted: #6B7C99;
    --bad: #FF5B5B;
    --warn: #F5A623;
    --good: #39D98A;
    --border: rgba(77,168,255,0.08);
    --radius: 7px;
  }

  * { margin: 0; padding: 0; box-sizing: border-box; }

  body {
    background: var(--bg);
    color: var(--text);
    font-family: 'DM Sans', sans-serif;
    display: flex;
    height: 100vh;
    overflow: hidden;
  }

  .sidebar {
    width: 220px;
    min-width: 220px;
    background: var(--sidebar);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
  }

  .sidebar-logo {
    padding: 20px 20px 16px;
    border-bottom: 1px solid var(--border);
  }

  .logo-mark {
    font-family: 'JetBrains Mono', monospace;
    font-size: 1.3rem;
    font-weight: 700;
    color: var(--accent);
    letter-spacing: -0.02em;
    line-height: 1;
  }

  .logo-sub {
    font-size: 0.7rem;
    color: var(--muted);
    margin-top: 3px;
    font-family: 'DM Sans', sans-serif;
    letter-spacing: 0.05em;
    text-transform: uppercase;
  }

  .sidebar-nav {
    flex: 1;
    padding: 12px 10px;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 12px;
    border-radius: var(--radius);
    cursor: pointer;
    font-size: 0.85rem;
    color: var(--muted);
    transition: all 0.15s;
    font-family: 'DM Sans', sans-serif;
    border: none;
    background: none;
    width: 100%;
    text-align: left;
  }

  .nav-item:hover { background: rgba(77,168,255,0.06); color: var(--text); }
  .nav-item.active { background: rgba(77,168,255,0.12); color: var(--accent); }

  .nav-badge {
    margin-left: auto;
    color: white;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.65rem;
    padding: 1px 6px;
    border-radius: 10px;
    display: none;
  }

  .sidebar-footer {
    padding: 14px 20px;
    border-top: 1px solid var(--border);
    font-size: 0.72rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
  }

  .status-dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    background: var(--good);
    margin-right: 6px;
    animation: pulse 2s infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.4; }
  }

  .main {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .page { display: none; }
  .page.active { display: contents; }

  .page-header {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    color: var(--muted);
    letter-spacing: 0.08em;
    text-transform: uppercase;
    margin-bottom: 4px;
  }

  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .grid-3 { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 16px; }

  .card {
    background: var(--card);
    border-radius: var(--radius);
    border: 1px solid var(--border);
    padding: 18px 20px;
  }

  .card-title {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    color: var(--accent);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    margin-bottom: 14px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .card-title::before {
    content: '';
    display: inline-block;
    width: 3px;
    height: 12px;
    background: var(--accent);
    border-radius: 2px;
  }

  .stat-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 5px 0;
    border-bottom: 1px solid rgba(255,255,255,0.03);
  }

  .stat-row:last-child { border-bottom: none; }

  .stat-label {
    font-size: 0.8rem;
    color: var(--muted);
    font-family: 'DM Sans', sans-serif;
  }

  .stat-value {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.82rem;
    color: var(--text);
  }

  .progress-wrap { margin-top: 3px; }

  .progress-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }

  .progress-bar {
    width: 100%;
    height: 4px;
    background: rgba(255,255,255,0.06);
    border-radius: 2px;
    overflow: hidden;
    margin-bottom: 10px;
  }

  .progress-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.5s ease;
  }

  .good-fill { background: var(--good); }
  .warn-fill { background: var(--warn); }
  .bad-fill { background: var(--bad); }

  table { width: 100%; border-collapse: collapse; }

  th {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--muted);
    text-align: left;
    padding: 6px 10px;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    border-bottom: 1px solid var(--border);
  }

  td {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.78rem;
    color: var(--text);
    padding: 6px 10px;
    border-bottom: 1px solid rgba(255,255,255,0.02);
  }

  tr:last-child td { border-bottom: none; }
  tr:hover td { background: rgba(77,168,255,0.04); }

  /* OVERVIEW */
  .overview-row { display: flex; gap: 16px; }
  .overview-row .card { flex: 1; }
  .alert-list, .suggestion-list { display: flex; flex-direction: column; gap: 8px; }

  .compact-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 10px 12px;
    background: var(--bg);
    border-radius: var(--radius);
    border-left: 3px solid var(--muted);
    cursor: pointer;
    transition: background 0.15s;
  }

  .compact-item:hover { background: rgba(77,168,255,0.04); }
  .compact-item.critical { border-left-color: var(--bad); }
  .compact-item.warning { border-left-color: var(--warn); }
  .compact-item.info { border-left-color: var(--accent); }
  .compact-item.suggestion { border-left-color: var(--good); }

  .compact-icon {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    min-width: 20px;
    margin-top: 1px;
  }

  .compact-item.critical .compact-icon { color: var(--bad); }
  .compact-item.warning .compact-icon { color: var(--warn); }
  .compact-item.info .compact-icon { color: var(--accent); }
  .compact-item.suggestion .compact-icon { color: var(--good); }

  .compact-body { flex: 1; }

  .compact-title {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    font-weight: 600;
    margin-bottom: 2px;
  }

  .compact-item.critical .compact-title { color: var(--bad); }
  .compact-item.warning .compact-title { color: var(--warn); }
  .compact-item.info .compact-title { color: var(--accent); }
  .compact-item.suggestion .compact-title { color: var(--good); }

  .compact-desc {
    font-size: 0.78rem;
    color: var(--muted);
    font-family: 'DM Sans', sans-serif;
    line-height: 1.4;
  }

  .compact-action {
    margin-top: 6px;
    background: none;
    border: 1px solid currentColor;
    border-radius: 4px;
    padding: 3px 10px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.68rem;
    cursor: pointer;
    transition: opacity 0.15s;
    display: inline-block;
  }

  .compact-item.critical .compact-action { color: var(--bad); }
  .compact-item.warning .compact-action { color: var(--warn); }
  .compact-item.info .compact-action { color: var(--accent); }
  .compact-item.suggestion .compact-action { color: var(--good); }
  .compact-action:hover { opacity: 0.7; }

  .empty-state {
    font-size: 0.8rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
    padding: 8px 0;
  }

  /* ALERTS */
  .alert {
    background: var(--card);
    border-radius: var(--radius);
    border: 1px solid var(--border);
    border-left: 3px solid var(--muted);
    padding: 16px 18px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .alert.critical { border-left-color: var(--bad); }
  .alert.warning { border-left-color: var(--warn); }
  .alert.info { border-left-color: var(--accent); }

  .alert-header { display: flex; align-items: center; gap: 10px; }

  .alert-icon { font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; font-weight: 700; }
  .alert.critical .alert-icon { color: var(--bad); }
  .alert.warning .alert-icon { color: var(--warn); }
  .alert.info .alert-icon { color: var(--accent); }

  .alert-title { font-family: 'JetBrains Mono', monospace; font-size: 0.78rem; font-weight: 600; letter-spacing: 0.05em; }
  .alert.critical .alert-title { color: var(--bad); }
  .alert.warning .alert-title { color: var(--warn); }
  .alert.info .alert-title { color: var(--accent); }

  .alert-desc { font-size: 0.82rem; color: var(--muted); font-family: 'DM Sans', sans-serif; line-height: 1.5; }

  .alert-action {
    margin-top: 4px;
    background: none;
    border: 1px solid currentColor;
    border-radius: 4px;
    padding: 5px 14px;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    cursor: pointer;
    align-self: flex-start;
    transition: opacity 0.15s;
  }

  .alert.critical .alert-action { color: var(--bad); }
  .alert.warning .alert-action { color: var(--warn); }
  .alert.info .alert-action { color: var(--accent); }
  .alert-action:hover { opacity: 0.7; }

  .maint-toolbar { display: flex; justify-content: space-between; align-items: center; }

  .maint-refresh {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    color: var(--muted);
    background: none;
    border: 1px solid var(--border);
    border-radius: 4px;
    padding: 4px 10px;
    cursor: pointer;
    transition: all 0.15s;
  }

  .maint-refresh:hover { color: var(--accent); border-color: var(--accent); }

  .maint-section-title {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.68rem;
    color: var(--muted);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    margin: 16px 0 8px;
  }

  /* ANALYSIS */
  .trend-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 20px;
    gap: 8px;
  }

  .trend-label {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: var(--muted);
    letter-spacing: 0.1em;
    text-transform: uppercase;
  }

  .trend-value {
    font-family: 'JetBrains Mono', monospace;
    font-size: 1.4rem;
    font-weight: 700;
  }

  .trend-up { color: var(--bad); }
  .trend-down { color: var(--good); }
  .trend-stable { color: var(--accent); }

  .trend-desc {
    font-size: 0.75rem;
    color: var(--muted);
    font-family: 'DM Sans', sans-serif;
  }

  .finding {
    background: var(--bg);
    border-radius: var(--radius);
    border-left: 3px solid var(--muted);
    padding: 14px 16px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .finding.critical { border-left-color: var(--bad); }
  .finding.warning { border-left-color: var(--warn); }
  .finding.info { border-left-color: var(--accent); }

  .finding-title {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    font-weight: 600;
  }

  .finding.critical .finding-title { color: var(--bad); }
  .finding.warning .finding-title { color: var(--warn); }
  .finding.info .finding-title { color: var(--accent); }

  .finding-detail {
    font-size: 0.8rem;
    color: var(--muted);
    font-family: 'DM Sans', sans-serif;
    line-height: 1.4;
  }

  .waiting-state {
    font-size: 0.82rem;
    color: var(--muted);
    font-family: 'JetBrains Mono', monospace;
    padding: 12px 0;
    text-align: center;
  }

  .settings-group { margin-bottom: 24px; }

  .settings-label {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    color: var(--accent);
    letter-spacing: 0.1em;
    text-transform: uppercase;
    margin-bottom: 12px;
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .settings-label::before {
    content: '';
    display: inline-block;
    width: 3px;
    height: 12px;
    background: var(--accent);
    border-radius: 2px;
  }
</style>
</head>
<body>

<div class="sidebar">
  <div class="sidebar-logo">
    <div class="logo-mark">sysix</div>
    <div class="logo-sub">by System9</div>
  </div>
  <nav class="sidebar-nav">
    <button class="nav-item active" onclick="showPage('overview', this)">
      Overview
    </button>
    <button class="nav-item" onclick="showPage('processes', this)">
      Processes
    </button>
    <button class="nav-item" onclick="showPage('ports', this)">
      Ports
    </button>
    <button class="nav-item" onclick="showPage('network', this)">
      Network
    </button>
    <button class="nav-item" onclick="showPage('system', this)">
      System
    </button>
    <button class="nav-item" onclick="showPage('analysis', this)">
      Analysis
      <span class="nav-badge" id="analysis-badge">0</span>
    </button>
    <button class="nav-item" id="nav-maintenance" onclick="showPage('maintenance', this)">
      Maintenance
      <span class="nav-badge" id="maint-badge">0</span>
    </button>
    <button class="nav-item" onclick="showPage('settings', this)">
      Settings
    </button>
  </nav>
  <div class="sidebar-footer">
    <span class="status-dot"></span>live
  </div>
</div>

<main class="main">

  <!-- OVERVIEW -->
  <div class="page active" id="page-overview">
    <div class="page-header">Overview</div>
    <div class="grid-2">
      <div class="card">
        <div class="card-title">System</div>
        <div class="stat-row"><span class="stat-label">Host</span><span class="stat-value" id="host">—</span></div>
        <div class="stat-row"><span class="stat-label">OS</span><span class="stat-value" id="os">—</span></div>
        <div class="stat-row"><span class="stat-label">Uptime</span><span class="stat-value" id="uptime">—</span></div>
      </div>
      <div class="card">
        <div class="card-title">Network</div>
        <div class="stat-row"><span class="stat-label">Sent</span><span class="stat-value" id="sent">—</span></div>
        <div class="stat-row"><span class="stat-label">Received</span><span class="stat-value" id="recv">—</span></div>
        <div class="stat-row"><span class="stat-label">Pkts Out</span><span class="stat-value" id="pkts-out">—</span></div>
        <div class="stat-row"><span class="stat-label">Pkts In</span><span class="stat-value" id="pkts-in">—</span></div>
      </div>
    </div>
    <div class="card">
      <div class="card-title">Resources</div>
      <div class="progress-wrap">
        <div class="progress-row"><span class="stat-label">CPU</span><span class="stat-value" id="cpu-val">—</span></div>
        <div class="progress-bar"><div class="progress-fill" id="cpu-bar"></div></div>
        <div class="progress-row"><span class="stat-label">Memory</span><span class="stat-value" id="mem-val">—</span></div>
        <div class="progress-bar"><div class="progress-fill" id="mem-bar"></div></div>
        <div class="progress-row"><span class="stat-label">Disk</span><span class="stat-value" id="disk-val">—</span></div>
        <div class="progress-bar"><div class="progress-fill" id="disk-bar" style="margin-bottom:0"></div></div>
      </div>
    </div>
    <div class="overview-row">
      <div class="card">
        <div class="card-title">Alerts</div>
        <div class="alert-list" id="overview-alerts">
          <div class="empty-state">Analyzing...</div>
        </div>
      </div>
      <div class="card">
        <div class="card-title">Suggestions</div>
        <div class="suggestion-list" id="overview-suggestions">
          <div class="empty-state">Loading...</div>
        </div>
      </div>
    </div>
  </div>

  <!-- PROCESSES -->
  <div class="page" id="page-processes">
    <div class="page-header">Processes</div>
    <div class="card">
      <div class="card-title">Running Processes</div>
      <table>
        <thead><tr><th>PID</th><th>Name</th><th>CPU%</th><th>Mem</th></tr></thead>
        <tbody id="procs"></tbody>
      </table>
    </div>
  </div>

  <!-- PORTS -->
  <div class="page" id="page-ports">
    <div class="page-header">Ports</div>
    <div class="card">
      <div class="card-title">Open Ports</div>
      <table>
        <thead><tr><th>Port</th><th>Type</th><th>Status</th><th>PID</th></tr></thead>
        <tbody id="ports"></tbody>
      </table>
    </div>
  </div>

  <!-- NETWORK -->
  <div class="page" id="page-network">
    <div class="page-header">Network</div>
    <div class="card">
      <div class="card-title">I/O Statistics</div>
      <div class="stat-row"><span class="stat-label">Download Speed</span><span class="stat-value" id="net-down-speed">—</span></div>
      <div class="stat-row"><span class="stat-label">Upload Speed</span><span class="stat-value" id="net-up-speed">—</span></div>
      <div class="stat-row"><span class="stat-label">Total Sent</span><span class="stat-value" id="net-sent">—</span></div>
      <div class="stat-row"><span class="stat-label">Total Received</span><span class="stat-value" id="net-recv">—</span></div>
      <div class="stat-row"><span class="stat-label">Packets Out</span><span class="stat-value" id="net-pkts-out">—</span></div>
      <div class="stat-row"><span class="stat-label">Packets In</span><span class="stat-value" id="net-pkts-in">—</span></div>
    </div>
  </div>

  <!-- SYSTEM -->
  <div class="page" id="page-system">
    <div class="page-header">System</div>
    <div class="card" style="margin-bottom:16px">
      <div class="card-title">CPU Usage</div>
      <canvas id="cpu-graph" height="80" style="width:100%;display:block;"></canvas>
    </div>
    <div class="card" style="margin-bottom:16px">
      <div class="card-title">Memory Usage</div>
      <canvas id="mem-graph" height="80" style="width:100%;display:block;"></canvas>
    </div>
    <div class="card" style="margin-bottom:16px">
      <div class="card-title">Disk Usage</div>
      <canvas id="disk-graph" height="80" style="width:100%;display:block;"></canvas>
    </div>
    <div class="grid-2">
      <div class="card">
        <div class="card-title">Details</div>
        <div class="stat-row"><span class="stat-label">Host</span><span class="stat-value" id="sys-host">—</span></div>
        <div class="stat-row"><span class="stat-label">OS</span><span class="stat-value" id="sys-os">—</span></div>
        <div class="stat-row"><span class="stat-label">Uptime</span><span class="stat-value" id="sys-uptime">—</span></div>
        <div class="stat-row"><span class="stat-label">CPU</span><span class="stat-value" id="sys-cpu">—</span></div>
        <div class="stat-row"><span class="stat-label">Memory</span><span class="stat-value" id="sys-mem">—</span></div>
        <div class="stat-row"><span class="stat-label">Disk</span><span class="stat-value" id="sys-disk">—</span></div>
      </div>
      <div class="card">
        <div class="card-title">Health</div>
        <div class="stat-row"><span class="stat-label">CPU</span><span class="stat-value" id="health-cpu">—</span></div>
        <div class="stat-row"><span class="stat-label">Memory</span><span class="stat-value" id="health-mem">—</span></div>
        <div class="stat-row"><span class="stat-label">Disk</span><span class="stat-value" id="health-disk">—</span></div>
      </div>
    </div>
  </div>

  <!-- ANALYSIS -->
  <div class="page" id="page-analysis">
    <div class="maint-toolbar">
      <div class="page-header">Analysis</div>
      <button class="maint-refresh" onclick="refreshAnalysis()">[ refresh ]</button>
    </div>
    <div class="grid-3" style="margin-top:4px;">
      <div class="card trend-card">
        <div class="trend-label">CPU Trend</div>
        <div class="trend-value" id="trend-cpu">—</div>
        <div class="trend-desc" id="trend-cpu-desc">collecting data...</div>
      </div>
      <div class="card trend-card">
        <div class="trend-label">Memory Trend</div>
        <div class="trend-value" id="trend-mem">—</div>
        <div class="trend-desc" id="trend-mem-desc">collecting data...</div>
      </div>
      <div class="card trend-card">
        <div class="trend-label">Disk Trend</div>
        <div class="trend-value" id="trend-disk">—</div>
        <div class="trend-desc" id="trend-disk-desc">collecting data...</div>
      </div>
    </div>
    <div class="card" style="margin-top:16px;">
      <div class="card-title">Findings</div>
      <div id="analysis-findings" style="display:flex;flex-direction:column;gap:10px;margin-top:4px;">
        <div class="waiting-state">[~] collecting history — findings appear after ~20 seconds</div>
      </div>
    </div>
  </div>

  <!-- MAINTENANCE -->
  <div class="page" id="page-maintenance">
    <div class="maint-toolbar">
      <div class="page-header">Maintenance</div>
      <button class="maint-refresh" onclick="runMaintenance()">[ refresh ]</button>
    </div>
    <div class="maint-section-title">System Alerts</div>
    <div id="maint-alerts" style="display:flex;flex-direction:column;gap:10px;">
      <div class="alert info">
        <div class="alert-header">
          <span class="alert-icon">[i]</span>
          <span class="alert-title">LOADING</span>
        </div>
        <div class="alert-desc">Analyzing system...</div>
      </div>
    </div>
    <div class="maint-section-title">Analyzer Findings</div>
    <div id="maint-findings" style="display:flex;flex-direction:column;gap:10px;">
      <div class="waiting-state">[~] collecting history...</div>
    </div>
  </div>

  <!-- SETTINGS -->
  <div class="page" id="page-settings">
    <div class="page-header">Settings</div>
    <div class="card">
      <div class="settings-group">
        <div class="settings-label">About</div>
        <div class="stat-row"><span class="stat-label">Version</span><span class="stat-value">0.3</span></div>
        <div class="stat-row"><span class="stat-label">License</span><span class="stat-value">Apache 2.0</span></div>
        <div class="stat-row"><span class="stat-label">Author</span><span class="stat-value">System9 Software</span></div>
      </div>
    </div>
  </div>

</main>

<script>
function showPage(name, el) {
  document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
  document.querySelectorAll('.nav-item').forEach(n => n.classList.remove('active'));
  document.getElementById('page-' + name).classList.add('active');
  el.classList.add('active');
  if (name === 'system') {
    // Re-render after the page becomes visible so canvas width is valid.
    setTimeout(refreshSystem, 0);
  }
}

function fillClass(pct) {
  if (pct >= 90) return 'bad-fill';
  if (pct >= 70) return 'warn-fill';
  return 'good-fill';
}

function formatBytes(b) {
  if (b >= 1e9) return (b/1e9).toFixed(1) + ' GB';
  if (b >= 1e6) return (b/1e6).toFixed(1) + ' MB';
  if (b >= 1e3) return (b/1e3).toFixed(1) + ' KB';
  return b + ' B';
}

function formatRate(bytesPerSecond) {
  return formatBytes(bytesPerSecond) + '/s';
}

function setBar(id, pct) {
  const el = document.getElementById(id);
  el.style.width = pct + '%';
  el.className = 'progress-fill ' + fillClass(pct);
}

function navTo(page) {
  const btn = document.querySelector('[onclick*="showPage(\'' + page + '\'"]');
  if (btn) showPage(page, btn);
}

const SUGGESTIONS = [
  { title: 'NEW IN SYSIX', desc: 'sysix has been updated to version 0.3. Changelog coming soon.', action: 'View Changelog', onclick: null },
  { title: 'TIP: TERMINAL MODE', desc: 'Run "sysix watch" in your terminal for a live TUI dashboard without the browser.', action: null, onclick: null },
  { title: 'TIP: QUICK SNAPSHOT', desc: 'Run "sysix status --procs --ports" for a fast terminal snapshot including processes and ports.', action: null, onclick: null },
  { title: 'CONFIG', desc: 'Customize refresh rate and visible panels in config.yaml in your sysix directory.', action: 'View Settings', onclick: 'settings' }
];

function compactItem(level, icon, title, desc, action, onclick) {
  const actionBtn = action
    ? '<button class="compact-action" onclick="' + (onclick ? 'navTo(\'' + onclick + '\')' : '') + '">' + action + ' →</button>'
    : '';
  return '<div class="compact-item ' + level + '" ' + (onclick && !action ? 'onclick="navTo(\'' + onclick + '\')"' : '') + '>' +
    '<span class="compact-icon">' + icon + '</span>' +
    '<div class="compact-body"><div class="compact-title">' + title + '</div><div class="compact-desc">' + desc + '</div>' + actionBtn + '</div>' +
    '</div>';
}

function trendIcon(t) {
  if (t === 'up') return '↑';
  if (t === 'down') return '↓';
  return '→';
}

function trendClass(t) {
  if (t === 'up') return 'trend-up';
  if (t === 'down') return 'trend-down';
  return 'trend-stable';
}

function trendDesc(t, label) {
  if (t === 'up') return label + ' is trending upward';
  if (t === 'down') return label + ' is trending downward';
  return label + ' is stable';
}

function renderFindings(findings, containerId) {
  const container = document.getElementById(containerId);
  if (!findings || findings.length === 0) {
    container.innerHTML = '<div class="finding info"><div class="finding-title">NO FINDINGS</div><div class="finding-detail">No anomalies detected in the current observation window.</div></div>';
    return;
  }
  container.innerHTML = findings.map(f => {
    const icons = { critical: '[!]', warning: '[~]', info: '[i]' };
    return '<div class="finding ' + f.Level + '">' +
      '<div class="finding-title">' + (icons[f.Level] || '[i]') + ' ' + f.Title + '</div>' +
      '<div class="finding-detail">' + f.Detail + '</div>' +
      '</div>';
  }).join('');
}

async function refreshAnalysis() {
  try {
    const report = await fetch('/api/analysis').then(r => r.json());

    // Trend cards
    ['cpu', 'mem', 'disk'].forEach((key, i) => {
      const labels = ['CPU', 'Memory', 'Disk'];
      const trendKeys = ['CPUTrend', 'MemTrend', 'DiskTrend'];
      const t = report[trendKeys[i]];
      const el = document.getElementById('trend-' + key);
      const desc = document.getElementById('trend-' + key + '-desc');
      el.textContent = trendIcon(t);
      el.className = 'trend-value ' + trendClass(t);
      desc.textContent = trendDesc(t, labels[i]);
    });

    // Findings
    renderFindings(report.Findings, 'analysis-findings');
    renderFindings(report.Findings, 'maint-findings');

    // Badge on analysis nav item
    const badge = document.getElementById('analysis-badge');
    const serious = (report.Findings || []).filter(f => f.Level === 'critical' || f.Level === 'warning');
    if (serious.length > 0) {
      badge.style.display = 'inline-block';
      badge.style.background = serious.some(f => f.Level === 'critical') ? 'var(--bad)' : 'var(--warn)';
      badge.textContent = serious.length;
    } else {
      badge.style.display = 'none';
    }

  } catch(e) { console.error(e); }
}

function analyzeSystem(snap, ports) {
  const alerts = [];
  if (snap.CPUPercent >= 90) {
    alerts.push({ level: 'critical', icon: '[!]', title: 'HIGH CPU USAGE', short: 'CPU at ' + snap.CPUPercent.toFixed(1) + '% — investigate processes.', desc: 'CPU is at ' + snap.CPUPercent.toFixed(1) + '%. Identify and investigate processes consuming excess cycles.', action: 'View Processes', page: 'processes' });
  } else if (snap.CPUPercent >= 70) {
    alerts.push({ level: 'warning', icon: '[~]', title: 'ELEVATED CPU USAGE', short: 'CPU at ' + snap.CPUPercent.toFixed(1) + '% — monitor for sustained load.', desc: 'CPU is at ' + snap.CPUPercent.toFixed(1) + '%. Monitor for sustained high load.', action: 'View Processes', page: 'processes' });
  }
  if (snap.MemPercent >= 90) {
    alerts.push({ level: 'critical', icon: '[!]', title: 'CRITICAL MEMORY PRESSURE', short: 'Memory at ' + snap.MemPercent.toFixed(1) + '% — action required.', desc: 'Memory is at ' + snap.MemPercent.toFixed(1) + '%. Identify and restart high-consumption processes immediately.', action: 'View Processes', page: 'processes' });
  } else if (snap.MemPercent >= 75) {
    alerts.push({ level: 'warning', icon: '[~]', title: 'HIGH MEMORY USAGE', short: 'Memory at ' + snap.MemPercent.toFixed(1) + '% — review services.', desc: 'Memory is at ' + snap.MemPercent.toFixed(1) + '%. Review running processes for unnecessary services.', action: 'View Processes', page: 'processes' });
  }
  if (snap.DiskPercent >= 90) {
    alerts.push({ level: 'critical', icon: '[!]', title: 'DISK SPACE CRITICAL', short: 'Disk at ' + snap.DiskPercent.toFixed(1) + '% — free space immediately.', desc: 'Disk is at ' + snap.DiskPercent.toFixed(1) + '%. Free up space immediately.', action: null, page: null });
  } else if (snap.DiskPercent >= 75) {
    alerts.push({ level: 'warning', icon: '[~]', title: 'DISK SPACE WARNING', short: 'Disk at ' + snap.DiskPercent.toFixed(1) + '% — plan for cleanup.', desc: 'Disk is at ' + snap.DiskPercent.toFixed(1) + '%. Plan for cleanup or expansion.', action: null, page: null });
  }
  const hours = Math.floor(snap.Uptime / 3600);
  if (hours >= 168) {
    alerts.push({ level: 'warning', icon: '[~]', title: 'EXTENDED UPTIME', short: 'System up ' + hours + ' hours — schedule maintenance window.', desc: 'System has been running for ' + hours + ' hours. Consider scheduling a maintenance window.', action: null, page: null });
  }
  if (ports && ports.length > 50) {
    alerts.push({ level: 'warning', icon: '[~]', title: 'HIGH PORT COUNT', short: ports.length + ' ports listening — review for unexpected services.', desc: ports.length + ' ports are currently listening. Review for unexpected services.', action: 'View Ports', page: 'ports' });
  }
  return alerts;
}

function renderOverview(alerts) {
  const alertContainer = document.getElementById('overview-alerts');
  const suggContainer = document.getElementById('overview-suggestions');
  const badge = document.getElementById('maint-badge');
  const urgent = alerts.filter(a => a.level === 'critical' || a.level === 'warning');
  if (urgent.length > 0) {
    badge.style.display = 'inline-block';
    badge.textContent = urgent.length;
    badge.style.background = alerts.some(a => a.level === 'critical') ? 'var(--bad)' : 'var(--warn)';
  } else {
    badge.style.display = 'none';
  }
  if (alerts.length === 0) {
    alertContainer.innerHTML = compactItem('info', '[i]', 'NO ALERTS', 'No alerts to display. System is healthy.', null, null);
  } else {
    alertContainer.innerHTML = alerts.map(a => compactItem(a.level, a.icon, a.title, a.short, a.action, a.page)).join('');
  }
  suggContainer.innerHTML = SUGGESTIONS.map(s => compactItem('suggestion', '[i]', s.title, s.desc, s.action, s.onclick)).join('');
}

function renderMaintenanceAlerts(alerts) {
  const container = document.getElementById('maint-alerts');
  if (alerts.length === 0) {
    container.innerHTML = '<div class="alert info"><div class="alert-header"><span class="alert-icon">[i]</span><span class="alert-title">ALL CLEAR</span></div><div class="alert-desc">No alerts detected. System is operating within normal parameters.</div></div>';
    return;
  }
  container.innerHTML = alerts.map(a => {
    const actionBtn = a.action ? '<button class="alert-action" onclick="navTo(\'' + a.page + '\')">' + a.action + ' →</button>' : '';
    return '<div class="alert ' + a.level + '"><div class="alert-header"><span class="alert-icon">' + a.icon + '</span><span class="alert-title">' + a.title + '</span></div><div class="alert-desc">' + a.desc + '</div>' + actionBtn + '</div>';
  }).join('');
}

async function runMaintenance() {
  try {
    const [snap, ports] = await Promise.all([
      fetch('/api/snapshot').then(r => r.json()),
      fetch('/api/ports').then(r => r.json()),
    ]);
    const alerts = analyzeSystem(snap, ports);
    renderOverview(alerts);
    renderMaintenanceAlerts(alerts);
  } catch(e) { console.error(e); }
}

let prevNetworkSample = null;

async function refresh() {
  try {
    const [snap, net, ports] = await Promise.all([
      fetch('/api/snapshot').then(r => r.json()),
      fetch('/api/network').then(r => r.json()),
      fetch('/api/ports').then(r => r.json()),
    ]);
    document.getElementById('host').textContent = snap.Hostname;
    document.getElementById('os').textContent = snap.OS;
    document.getElementById('uptime').textContent = Math.floor(snap.Uptime / 3600) + ' hours';
    document.getElementById('cpu-val').textContent = snap.CPUPercent.toFixed(1) + '%';
    setBar('cpu-bar', snap.CPUPercent);
    document.getElementById('mem-val').textContent = snap.MemPercent.toFixed(1) + '% (' + Math.floor(snap.MemUsed/1024/1024) + ' MB / ' + Math.floor(snap.MemTotal/1024/1024) + ' MB)';
    setBar('mem-bar', snap.MemPercent);
    document.getElementById('disk-val').textContent = snap.DiskPercent.toFixed(1) + '% (' + Math.floor(snap.DiskUsed/1024/1024/1024) + ' GB / ' + Math.floor(snap.DiskTotal/1024/1024/1024) + ' GB)';
    setBar('disk-bar', snap.DiskPercent);
    document.getElementById('sent').textContent = formatBytes(net.BytesSent);
    document.getElementById('recv').textContent = formatBytes(net.BytesRecv);
    document.getElementById('pkts-out').textContent = net.PacketsSent;
    document.getElementById('pkts-in').textContent = net.PacketsRecv;
    document.getElementById('net-sent').textContent = formatBytes(net.BytesSent);
    document.getElementById('net-recv').textContent = formatBytes(net.BytesRecv);
    document.getElementById('net-pkts-out').textContent = net.PacketsSent;
    document.getElementById('net-pkts-in').textContent = net.PacketsRecv;
    const now = Date.now();
    if (prevNetworkSample) {
      const elapsedSec = Math.max((now - prevNetworkSample.ts) / 1000, 0.001);
      const upBps = Math.max((net.BytesSent - prevNetworkSample.bytesSent) / elapsedSec, 0);
      const downBps = Math.max((net.BytesRecv - prevNetworkSample.bytesRecv) / elapsedSec, 0);
      document.getElementById('net-up-speed').textContent = formatRate(upBps);
      document.getElementById('net-down-speed').textContent = formatRate(downBps);
    } else {
      document.getElementById('net-up-speed').textContent = '—';
      document.getElementById('net-down-speed').textContent = '—';
    }
    prevNetworkSample = {
      bytesSent: net.BytesSent,
      bytesRecv: net.BytesRecv,
      ts: now,
    };
    document.getElementById('procs').innerHTML = snap.Processes.filter(p => p.MemMB > 1).slice(0, 20).map(p => '<tr><td>'+p.PID+'</td><td>'+p.Name+'</td><td>'+p.CPUPercent.toFixed(1)+'%</td><td>'+p.MemMB.toFixed(0)+' MB</td></tr>').join('');
    document.getElementById('ports').innerHTML = ports.slice(0, 20).map(p => '<tr><td>'+p.Port+'</td><td>'+p.Type+'</td><td>'+p.Status+'</td><td>'+p.PID+'</td></tr>').join('');
    const alerts = analyzeSystem(snap, ports);
    renderOverview(alerts);
    renderMaintenanceAlerts(alerts);
  } catch(e) { console.error(e); }
}

function drawGraph(canvasId, data, color) {
  const canvas = document.getElementById(canvasId);
  if (!canvas) return;
  canvas.width = canvas.offsetWidth;
  const ctx = canvas.getContext('2d');
  const w = canvas.width, h = canvas.height;
  ctx.clearRect(0, 0, w, h);
  if (data.length < 2) return;
  ctx.beginPath();
  ctx.strokeStyle = color;
  ctx.lineWidth = 2;
  ctx.lineJoin = 'round';
  data.forEach((val, i) => {
    const x = (i / (data.length - 1)) * w;
    const y = h - (val / 100) * h;
    i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y);
  });
  ctx.stroke();
  ctx.lineTo(w, h); ctx.lineTo(0, h); ctx.closePath();
  ctx.fillStyle = color + '22';
  ctx.fill();
}

function healthText(pct) {
  if (pct >= 90) return '<span style="color:var(--bad)">Critical (' + pct.toFixed(1) + '%)</span>';
  if (pct >= 70) return '<span style="color:var(--warn)">Warning (' + pct.toFixed(1) + '%)</span>';
  return '<span style="color:var(--good)">Healthy (' + pct.toFixed(1) + '%)</span>';
}

async function refreshSystem() {
  try {
    const [snap, hist] = await Promise.all([
      fetch('/api/snapshot').then(r => r.json()),
      fetch('/api/history').then(r => r.json()),
    ]);
    document.getElementById('sys-host').textContent = snap.Hostname;
    document.getElementById('sys-os').textContent = snap.OS;
    document.getElementById('sys-uptime').textContent = Math.floor(snap.Uptime / 3600) + ' hours';
    document.getElementById('sys-cpu').textContent = snap.CPUPercent.toFixed(1) + '%';
    document.getElementById('sys-mem').textContent = snap.MemPercent.toFixed(1) + '% (' + Math.floor(snap.MemUsed/1024/1024) + ' MB / ' + Math.floor(snap.MemTotal/1024/1024) + ' MB)';
    document.getElementById('sys-disk').textContent = snap.DiskPercent.toFixed(1) + '% (' + Math.floor(snap.DiskUsed/1024/1024/1024) + ' GB / ' + Math.floor(snap.DiskTotal/1024/1024/1024) + ' GB)';
    document.getElementById('health-cpu').innerHTML = healthText(snap.CPUPercent);
    document.getElementById('health-mem').innerHTML = healthText(snap.MemPercent);
    document.getElementById('health-disk').innerHTML = healthText(snap.DiskPercent);
    if (hist && hist.length > 1) {
      drawGraph('cpu-graph', hist.map(h => h.cpu), '#4DA8FF');
      drawGraph('mem-graph', hist.map(h => h.mem), '#39D98A');
      drawGraph('disk-graph', hist.map(h => h.disk), '#F5A623');
    }
  } catch(e) { console.error(e); }
}

refresh();
refreshSystem();
refreshAnalysis();
setInterval(refresh, 2000);
setInterval(refreshSystem, 2000);
setInterval(refreshAnalysis, 5000);
</script>
</body>
</html>`
