

// Detect base path for subpath deployments
function getBasePath() {
    let path = window.location.pathname;
    // Remove trailing filename if present
    if (path.endsWith('.html')) path = path.substring(0, path.lastIndexOf('/') + 1);
    // Ensure trailing slash
    if (!path.endsWith('/')) path += '/';
    return path;
}
const BASE = getBasePath();

async function loadHosts() {
    const res = await fetch(BASE + 'api/hosts');
    const hosts = await res.json();
    const hostsDiv = document.getElementById('hosts');
    hostsDiv.innerHTML = '<ul>' + hosts.map(h => {
        const date = h.last_seen ? new Date(h.last_seen * 1000).toLocaleString() : 'Never';
        const now = Date.now() / 1000;
        let badgeColor = 'red';
        if (h.last_seen && (now - h.last_seen) < 600) badgeColor = 'green';
        const badge = `<span class="host-badge" style="display:inline-block;width:12px;height:12px;border-radius:50%;background:${badgeColor};vertical-align:middle;"></span>`;
        return `<li>${badge}<span class="host-name">${h.hostname}</span><span class="host-date">${date}</span></li>`;
    }).join('') + '</ul>';
}


window.onload = loadHosts;

