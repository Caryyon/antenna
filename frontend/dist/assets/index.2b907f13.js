(function(){const n=document.createElement("link").relList;if(n&&n.supports&&n.supports("modulepreload"))return;for(const a of document.querySelectorAll('link[rel="modulepreload"]'))i(a);new MutationObserver(a=>{for(const s of a)if(s.type==="childList")for(const o of s.addedNodes)o.tagName==="LINK"&&o.rel==="modulepreload"&&i(o)}).observe(document,{childList:!0,subtree:!0});function l(a){const s={};return a.integrity&&(s.integrity=a.integrity),a.referrerpolicy&&(s.referrerPolicy=a.referrerpolicy),a.crossorigin==="use-credentials"?s.credentials="include":a.crossorigin==="anonymous"?s.credentials="omit":s.credentials="same-origin",s}function i(a){if(a.ep)return;a.ep=!0;const s=l(a);fetch(a.href,s)}})();function d(){return window.go.main.App.GetDashboard()}const t=e=>`$${e.toFixed(2)}`;function r(e){const n=e.sessions.filter(s=>s.kind==="main"&&s.isActive),l=e.sessions.filter(s=>s.kind==="main"&&!s.isActive),i=e.sessions.filter(s=>s.kind==="subagent"),a=e.sessions.filter(s=>s.kind==="cron");document.getElementById("app").innerHTML=`
        <div class="dashboard">
            <!-- Stats Bar -->
            <div class="stats-bar">
                <div class="stat-group">
                    <span class="live-dot"></span>
                    <span class="label">Live</span>
                </div>
                <div class="stat-group">
                    <span class="stat-value big">${e.totalCount}</span>
                    <span class="label">sessions</span>
                </div>
                <div class="stat-group">
                    <span class="dot green"></span>
                    <span class="stat-value green">${n.length}</span>
                    <span class="label">active</span>
                </div>
                <div class="stat-group">
                    <span class="dot purple"></span>
                    <span class="stat-value purple">${i.length}</span>
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
                    <div class="cost-value green">${t(e.todayCost)}</div>
                </div>
                <div class="cost-group">
                    <div class="cost-label">Total</div>
                    <div class="cost-value">${t(e.totalCost)}</div>
                </div>
            </div>

            <!-- Main Grid -->
            <div class="grid">
                <!-- Left Panel -->
                <div class="left-panel">
                    ${n.length>0?`
                    <div class="section active-section">
                        <div class="section-header">
                            <span class="live-dot small"></span>
                            <span class="section-title green">Active</span>
                        </div>
                        <div class="rows">
                            ${n.map(s=>`
                            <div class="row">
                                <span class="session-name">${s.name}</span>
                                <span class="session-id">${s.sessionId}</span>
                                <span class="model">${s.model||""}</span>
                                <span class="msgs">${s.messageCount}</span>
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
                                <span class="session-name">${s.name}</span>
                                <span class="session-id">${s.sessionId}</span>
                                <span class="msgs">${s.messageCount}</span>
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
                            <span class="count purple">${i.length}</span>
                        </div>
                        <div class="rows scrollable">
                            ${i.length>0?i.map(s=>`
                            <div class="card">
                                <div class="card-header">
                                    <span class="card-name">${s.name}</span>
                                    ${s.isActive?'<span class="live-dot small"></span>':""}
                                </div>
                                <div class="card-meta">
                                    <span>${s.messageCount} msgs</span>
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
                                    <span class="card-name">${s.name}</span>
                                    ${s.isActive?'<span class="live-dot small"></span>':""}
                                </div>
                                <div class="card-meta">
                                    <span>${s.messageCount} msgs</span>
                                    <span>${t(s.totalCost)}</span>
                                </div>
                            </div>
                            `).join(""):'<div class="empty">No cron jobs</div>'}
                        </div>
                    </div>
                </div>
            </div>
        </div>
    `}async function c(){try{const e=await d();r(e)}catch(e){console.error("Failed to get dashboard:",e)}}c();setInterval(c,5e3);
