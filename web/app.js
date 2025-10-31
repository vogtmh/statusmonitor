
async function loadHosts() {
    const res = await fetch('/api/hosts');
    const hosts = await res.json();
    const hostsDiv = document.getElementById('hosts');
    hostsDiv.innerHTML = '<ul>' + hosts.map(h => {
        const date = h.last_seen ? new Date(h.last_seen * 1000).toLocaleString() : 'Never';
        const now = Date.now() / 1000;
        let badgeColor = 'red';
        if (h.last_seen && (now - h.last_seen) < 600) badgeColor = 'green';
        const badge = `<span style="display:inline-block;width:12px;height:12px;border-radius:50%;background:${badgeColor};margin-right:8px;vertical-align:middle;"></span>`;
        return `<li>${badge}<a href="#" onclick="showChart('${h.hostname}')">${h.hostname}</a> <span style="color:#888;font-size:0.9em;">(${date})</span></li>`;
    }).join('') + '</ul>';
}

async function showChart(hostname) {
    document.getElementById('chart-container').style.display = 'block';
    document.getElementById('host-title').textContent = hostname;
    const res = await fetch(`/api/history?hostname=${encodeURIComponent(hostname)}`);
    const history = await res.json();
    const labels = history.map(e => new Date(e.timestamp * 1000).toLocaleString());
    const data = history.map(e => e.up ? 1 : 0);
    if (window.statusChart && typeof window.statusChart.destroy === 'function') {
        window.statusChart.destroy();
    }
    const ctx = document.getElementById('statusChart').getContext('2d');
    window.statusChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels,
            datasets: [{
                label: 'Up (1) / Down (0)',
                data,
                borderColor: 'green',
                backgroundColor: 'rgba(0,255,0,0.1)',
                fill: true,
                stepped: true
            }]
        },
        options: {
            scales: {
                y: {
                    min: 0,
                    max: 1,
                    ticks: { stepSize: 1 }
                }
            }
        }
    });
}

function closeChart() {
    document.getElementById('chart-container').style.display = 'none';
}

window.onload = loadHosts;

