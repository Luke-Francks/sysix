package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/System9-Software/sysix/internal/collector"
)

func Start(port int) error {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/snapshot", handleSnapshot)
	http.HandleFunc("/api/ports", handlePorts)
	http.HandleFunc("/api/network", handleNetwork)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("sysix web UI running at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboard))
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

  /* SIDEBAR */
  .sidebar {
    width: 220px;
    min-width: 220px;
    background: var(--sidebar);
    border-right: 1px solid var(--border);
    display: flex;
    flex-direction: column;
    padding: 0;
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

  .nav-icon {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.75rem;
    width: 18px;
    opacity: 0.8;
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

  /* MAIN */
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

  /* CARDS */
  .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }
  .grid-1 { display: grid; grid-template-columns: 1fr; gap: 16px; }

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

  /* PROGRESS BARS */
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

  /* TABLES */
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

  /* SETTINGS */
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

  .logo-options {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 10px;
  }

  .logo-option {
    background: var(--card);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 14px 10px;
    cursor: pointer;
    text-align: center;
    transition: all 0.15s;
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.72rem;
    color: var(--muted);
  }

  .logo-option:hover { border-color: var(--accent); color: var(--text); }
  .logo-option.selected { border-color: var(--accent); background: rgba(77,168,255,0.08); color: var(--accent); }

  .logo-preview {
    font-size: 1.1rem;
    color: var(--accent);
    margin-bottom: 6px;
    font-weight: 700;
  }
</style>
</head>
<body>

<div class="sidebar">
  <div class="sidebar-logo">
    <div class="logo-mark" id="sidebar-logo">sysix</div>
    <div class="logo-sub">by System9</div>
  </div>
  <nav class="sidebar-nav">
    <button class="nav-item active" onclick="showPage('overview', this)">
      <span class="nav-icon">[*]</span> Overview
    </button>
    <button class="nav-item" onclick="showPage('processes', this)">
      <span class="nav-icon">[>]</span> Processes
    </button>
    <button class="nav-item" onclick="showPage('ports', this)">
      <span class="nav-icon">[~]</span> Ports
    </button>
    <button class="nav-item" onclick="showPage('network', this)">
      <span class="nav-icon">[-]</span> Network
    </button>
    <button class="nav-item" onclick="showPage('settings', this)">
      <span class="nav-icon">[+]</span> Settings
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
      <div class="stat-row"><span class="stat-label">Total Sent</span><span class="stat-value" id="net-sent">—</span></div>
      <div class="stat-row"><span class="stat-label">Total Received</span><span class="stat-value" id="net-recv">—</span></div>
      <div class="stat-row"><span class="stat-label">Packets Out</span><span class="stat-value" id="net-pkts-out">—</span></div>
      <div class="stat-row"><span class="stat-label">Packets In</span><span class="stat-value" id="net-pkts-in">—</span></div>
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

function setBar(id, pct) {
  const el = document.getElementById(id);
  el.style.width = pct + '%';
  el.className = 'progress-fill ' + fillClass(pct);
}

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

    document.getElementById('procs').innerHTML = snap.Processes
      .filter(p => p.MemMB > 1)
      .slice(0, 20)
      .map(p => '<tr><td>'+p.PID+'</td><td>'+p.Name+'</td><td>'+p.CPUPercent.toFixed(1)+'%</td><td>'+p.MemMB.toFixed(0)+' MB</td></tr>')
      .join('');

    document.getElementById('ports').innerHTML = ports
      .slice(0, 20)
      .map(p => '<tr><td>'+p.Port+'</td><td>'+p.Type+'</td><td>'+p.Status+'</td><td>'+p.PID+'</td></tr>')
      .join('');

  } catch(e) { console.error(e); }
}

refresh();
setInterval(refresh, 2000);
</script>
</body>
</html>`
