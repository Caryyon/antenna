import { GetDashboard } from '../wailsjs/go/main/App';

const formatCost = (cost) => {
    if (cost === undefined || cost === null) return '$0.00';
    return `$${cost.toFixed(2)}`;
};

function renderError(message) {
    document.getElementById('app').innerHTML = `
        <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; color: #666; font-family: 'JetBrains Mono', monospace;">
            <div style="font-size: 48px; margin-bottom: 20px;">📡</div>
            <div style="font-size: 14px; color: #888; margin-bottom: 10px;">Antenna</div>
            <div style="font-size: 12px; color: #ff6b35; max-width: 400px; text-align: center;">${message}</div>
            <div style="font-size: 11px; color: #444; margin-top: 20px;">
                Looking for: ~/.openclaw/agents/main/sessions/
            </div>
        </div>
    `;
}

function renderDashboard(data) {
    if (!data || !data.sessions) {
        renderError('No data received from backend');
        return;
    }

    const sessions = data.sessions || [];
    const active = sessions.filter(s => s.kind === 'main' && s.isActive);
    const idle = sessions.filter(s => s.kind === 'main' && !s.isActive);
    const subs = sessions.filter(s => s.kind === 'subagent');
    const crons = sessions.filter(s => s.kind === 'cron');

    if (sessions.length === 0) {
        document.getElementById('app').innerHTML = `
            <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; color: #666; font-family: 'JetBrains Mono', monospace;">
                <div style="font-size: 48px; margin-bottom: 20px;">📡</div>
                <div style="font-size: 14px; color: #888; margin-bottom: 10px;">Antenna</div>
                <div style="font-size: 12px; color: #555;">No sessions found</div>
                <div style="font-size: 11px; color: #444; margin-top: 20px;">
                    Looking in: ~/.openclaw/agents/main/sessions/
                </div>
            </div>
        `;
        return;
    }

    document.getElementById('app').innerHTML = `
        <div class="dashboard">
            <!-- Stats Bar -->
            <div class="stats-bar">
                <div class="stat-group">
                    <span class="live-dot"></span>
                    <span class="label">Live</span>
                </div>
                <div class="stat-group">
                    <span class="stat-value big">${data.totalCount || 0}</span>
                    <span class="label">sessions</span>
                </div>
                <div class="stat-group">
                    <span class="dot green"></span>
                    <span class="stat-value green">${active.length}</span>
                    <span class="label">active</span>
                </div>
                <div class="stat-group">
                    <span class="dot purple"></span>
                    <span class="stat-value purple">${subs.length}</span>
                    <span class="label">sub</span>
                </div>
                <div class="stat-group">
                    <span class="dot orange"></span>
                    <span class="stat-value orange">${crons.length}</span>
                    <span class="label">cron</span>
                </div>
                <div class="spacer"></div>
                <div class="cost-group">
                    <div class="cost-label">Today</div>
                    <div class="cost-value green">${formatCost(data.todayCost)}</div>
                </div>
                <div class="cost-group">
                    <div class="cost-label">Total</div>
                    <div class="cost-value">${formatCost(data.totalCost)}</div>
                </div>
            </div>

            <!-- Main Grid -->
            <div class="grid">
                <!-- Left Panel -->
                <div class="left-panel">
                    ${active.length > 0 ? `
                    <div class="section active-section">
                        <div class="section-header">
                            <span class="live-dot small"></span>
                            <span class="section-title green">Active</span>
                        </div>
                        <div class="rows">
                            ${active.map(s => `
                            <div class="row">
                                <span class="session-name">${s.name || 'unnamed'}</span>
                                <span class="session-id">${s.sessionId || ''}</span>
                                <span class="model">${s.model || ''}</span>
                                <span class="msgs">${s.messageCount || 0}</span>
                                <span class="cost green">${formatCost(s.todayCost)}</span>
                                <span class="cost">${formatCost(s.totalCost)}</span>
                            </div>
                            `).join('')}
                        </div>
                    </div>
                    ` : ''}
                    
                    <div class="section idle-section">
                        <div class="section-header">
                            <span class="idle-dot"></span>
                            <span class="section-title gray">Idle</span>
                            <span class="count">${idle.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${idle.map(s => `
                            <div class="row dim">
                                <span class="session-name">${s.name || 'unnamed'}</span>
                                <span class="session-id">${s.sessionId || ''}</span>
                                <span class="msgs">${s.messageCount || 0}</span>
                                <span class="cost">${formatCost(s.totalCost)}</span>
                            </div>
                            `).join('')}
                        </div>
                    </div>
                </div>

                <!-- Right Panel -->
                <div class="right-panel">
                    <div class="section sub-section">
                        <div class="section-header">
                            <span class="icon">⚡</span>
                            <span class="section-title purple">Sub-agents</span>
                            <span class="count purple">${subs.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${subs.length > 0 ? subs.map(s => `
                            <div class="card">
                                <div class="card-header">
                                    <span class="card-name">${s.name || 'unnamed'}</span>
                                    ${s.isActive ? '<span class="live-dot small"></span>' : ''}
                                </div>
                                <div class="card-meta">
                                    <span>${s.messageCount || 0} msgs</span>
                                    <span>${formatCost(s.totalCost)}</span>
                                </div>
                            </div>
                            `).join('') : '<div class="empty">No sub-agents</div>'}
                        </div>
                    </div>

                    <div class="section cron-section">
                        <div class="section-header">
                            <span class="icon">⏱</span>
                            <span class="section-title orange">Cron</span>
                            <span class="count orange">${crons.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${crons.length > 0 ? crons.map(s => `
                            <div class="card">
                                <div class="card-header">
                                    <span class="card-name">${s.name || 'unnamed'}</span>
                                    ${s.isActive ? '<span class="live-dot small"></span>' : ''}
                                </div>
                                <div class="card-meta">
                                    <span>${s.messageCount || 0} msgs</span>
                                    <span>${formatCost(s.totalCost)}</span>
                                </div>
                            </div>
                            `).join('') : '<div class="empty">No cron jobs</div>'}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `;
}

async function refresh() {
    try {
        const data = await GetDashboard();
        renderDashboard(data);
    } catch (e) {
        console.error('Failed to get dashboard:', e);
        renderError(`Error: ${e.message || e}`);
    }
}

// Initial load
refresh();

// Auto-refresh every 5 seconds
setInterval(refresh, 5000);
