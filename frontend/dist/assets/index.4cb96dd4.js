(function(){const i=document.createElement("link").relList;if(i&&i.supports&&i.supports("modulepreload"))return;for(const e of document.querySelectorAll('link[rel="modulepreload"]'))l(e);new MutationObserver(e=>{for(const a of e)if(a.type==="childList")for(const s of a.addedNodes)s.tagName==="LINK"&&s.rel==="modulepreload"&&l(s)}).observe(document,{childList:!0,subtree:!0});function o(e){const a={};return e.integrity&&(a.integrity=e.integrity),e.referrerpolicy&&(a.referrerPolicy=e.referrerpolicy),e.crossorigin==="use-credentials"?a.credentials="include":e.crossorigin==="anonymous"?a.credentials="omit":a.credentials="same-origin",a}function l(e){if(e.ep)return;e.ep=!0;const a=o(e);fetch(e.href,a)}})();function r(){return window.go.main.App.GetDashboard()}const t=n=>n==null?"$0.00":`$${n.toFixed(2)}`;function c(n){document.getElementById("app").innerHTML=`
        <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; color: #666; font-family: 'JetBrains Mono', monospace;">
            <div style="font-size: 48px; margin-bottom: 20px;">\u{1F4E1}</div>
            <div style="font-size: 14px; color: #888; margin-bottom: 10px;">Antenna</div>
            <div style="font-size: 12px; color: #ff6b35; max-width: 400px; text-align: center;">${n}</div>
            <div style="font-size: 11px; color: #444; margin-top: 20px;">
                Looking for: ~/.openclaw/agents/main/sessions/
            </div>
        </div>
    `}function p(n){if(!n||!n.sessions){c("No data received from backend");return}const i=n.sessions||[],o=i.filter(s=>s.kind==="main"&&s.isActive),l=i.filter(s=>s.kind==="main"&&!s.isActive),e=i.filter(s=>s.kind==="subagent"),a=i.filter(s=>s.kind==="cron");if(i.length===0){document.getElementById("app").innerHTML=`
            <div style="display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; color: #666; font-family: 'JetBrains Mono', monospace;">
                <div style="font-size: 48px; margin-bottom: 20px;">\u{1F4E1}</div>
                <div style="font-size: 14px; color: #888; margin-bottom: 10px;">Antenna</div>
                <div style="font-size: 12px; color: #555;">No sessions found</div>
                <div style="font-size: 11px; color: #444; margin-top: 20px;">
                    Looking in: ~/.openclaw/agents/main/sessions/
                </div>
            </div>
        `;return}document.getElementById("app").innerHTML=`
        <div class="dashboard">
            <!-- Stats Bar -->
            <div class="stats-bar">
                <div class="stat-group">
                    <span class="live-dot"></span>
                    <span class="label">Live</span>
                </div>
                <div class="stat-group">
                    <span class="stat-value big">${n.totalCount||0}</span>
                    <span class="label">sessions</span>
                </div>
                <div class="stat-group">
                    <span class="dot green"></span>
                    <span class="stat-value green">${o.length}</span>
                    <span class="label">active</span>
                </div>
                <div class="stat-group">
                    <span class="dot purple"></span>
                    <span class="stat-value purple">${e.length}</span>
                    <span class="label">sub</span>
                </div>
                <div class="stat-group">
                    <span class="dot orange"></span>
                    <span class="stat-value orange">${a.length}</span>
                    <span class="label">cron</span>
                </div>
                <div class="spacer"></div>
                <div class="cost-group">
                    <div class="cost-label">Today</div>
                    <div class="cost-value green">${t(n.todayCost)}</div>
                </div>
                <div class="cost-group">
                    <div class="cost-label">Total</div>
                    <div class="cost-value">${t(n.totalCost)}</div>
                </div>
            </div>

            <!-- Main Grid -->
            <div class="grid">
                <!-- Left Panel -->
                <div class="left-panel">
                    ${o.length>0?`
                    <div class="section active-section">
                        <div class="section-header">
                            <span class="live-dot small"></span>
                            <span class="section-title green">Active</span>
                        </div>
                        <div class="rows">
                            ${o.map(s=>`
                            <div class="row">
                                <span class="session-name">${s.name||"unnamed"}</span>
                                <span class="session-id">${s.sessionId||""}</span>
                                <span class="model">${s.model||""}</span>
                                <span class="msgs">${s.messageCount||0}</span>
                                <span class="cost green">${t(s.todayCost)}</span>
                                <span class="cost">${t(s.totalCost)}</span>
                            </div>
                            `).join("")}
                        </div>
                    </div>
                    `:""}
                    
                    <div class="section idle-section">
                        <div class="section-header">
                            <span class="idle-dot"></span>
                            <span class="section-title gray">Idle</span>
                            <span class="count">${l.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${l.map(s=>`
                            <div class="row dim">
                                <span class="session-name">${s.name||"unnamed"}</span>
                                <span class="session-id">${s.sessionId||""}</span>
                                <span class="msgs">${s.messageCount||0}</span>
                                <span class="cost">${t(s.totalCost)}</span>
                            </div>
                            `).join("")}
                        </div>
                    </div>
                </div>

                <!-- Right Panel -->
                <div class="right-panel">
                    <div class="section sub-section">
                        <div class="section-header">
                            <span class="icon">\u26A1</span>
                            <span class="section-title purple">Sub-agents</span>
                            <span class="count purple">${e.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${e.length>0?e.map(s=>`
                            <div class="card">
                                <div class="card-header">
                                    <span class="card-name">${s.name||"unnamed"}</span>
                                    ${s.isActive?'<span class="live-dot small"></span>':""}
                                </div>
                                <div class="card-meta">
                                    <span>${s.messageCount||0} msgs</span>
                                    <span>${t(s.totalCost)}</span>
                                </div>
                            </div>
                            `).join(""):'<div class="empty">No sub-agents</div>'}
                        </div>
                    </div>

                    <div class="section cron-section">
                        <div class="section-header">
                            <span class="icon">\u23F1</span>
                            <span class="section-title orange">Cron</span>
                            <span class="count orange">${a.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${a.length>0?a.map(s=>`
                            <div class="card">
                                <div class="card-header">
                                    <span class="card-name">${s.name||"unnamed"}</span>
                                    ${s.isActive?'<span class="live-dot small"></span>':""}
                                </div>
                                <div class="card-meta">
                                    <span>${s.messageCount||0} msgs</span>
                                    <span>${t(s.totalCost)}</span>
                                </div>
                            </div>
                            `).join(""):'<div class="empty">No cron jobs</div>'}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `}async function d(){try{const n=await r();p(n)}catch(n){console.error("Failed to get dashboard:",n),c(`Error: ${n.message||n}`)}}d();setInterval(d,5e3);
