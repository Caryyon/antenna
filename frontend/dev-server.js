#!/usr/bin/env node
// Dev API server — reads real OpenClaw session data for browser dev mode
const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = 5174;
const OPENCLAW_DIR = path.join(require('os').homedir(), '.openclaw');
const SESSIONS_DIR = path.join(OPENCLAW_DIR, 'agents', 'main', 'sessions');

function loadCronJobNames() {
  const names = {};
  try {
    const data = JSON.parse(fs.readFileSync(path.join(OPENCLAW_DIR, 'cron', 'jobs.json'), 'utf8'));
    for (const job of data.jobs || []) {
      names[job.id] = job.name;
    }
  } catch {}
  return names;
}

function parseKind(key) {
  const parts = key.split(':');
  if (parts.length >= 3) {
    if (parts[2] === 'cron') return 'cron';
    if (parts[2] === 'subagent') return 'subagent';
  }
  return 'main';
}

function loadSessions() {
  const cronNames = loadCronJobNames();
  let sessionMeta = {};
  try {
    sessionMeta = JSON.parse(fs.readFileSync(path.join(SESSIONS_DIR, 'sessions.json'), 'utf8'));
  } catch {}

  const metaByID = {};
  for (const [key, entry] of Object.entries(sessionMeta)) {
    metaByID[entry.sessionId] = { key, entry };
  }

  const files = fs.readdirSync(SESSIONS_DIR).filter(f => f.endsWith('.jsonl'));
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const todayMs = today.getTime();
  const sessions = [];

  for (const file of files) {
    const sessionID = file.replace('.jsonl', '');
    const stat = fs.statSync(path.join(SESSIONS_DIR, file));
    const updatedAt = stat.mtimeMs;
    const isActive = (Date.now() - updatedAt) < 30 * 60 * 1000;

    const s = {
      sessionId: sessionID,
      name: '',
      kind: 'main',
      model: '',
      messageCount: 0,
      totalCost: 0,
      todayCost: 0,
      updatedAt,
      isActive,
    };

    const meta = metaByID[sessionID];
    if (meta) {
      s.name = meta.entry.label || '';
      s.model = meta.entry.model || '';
      s.kind = parseKind(meta.key);
      if (meta.entry.updatedAt > 0) {
        s.updatedAt = meta.entry.updatedAt;
        s.isActive = (Date.now() - meta.entry.updatedAt) < 30 * 60 * 1000;
      }
      if (s.kind === 'cron' && !s.name) {
        const parts = meta.key.split(':');
        if (parts.length >= 4 && cronNames[parts[3]]) {
          s.name = cronNames[parts[3]];
        }
      }
    }

    if (!s.name) {
      if (s.kind === 'main') s.name = new Date(s.updatedAt).toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
      else if (s.kind === 'cron') s.name = 'cron-' + sessionID.slice(0, 8);
      else s.name = sessionID.slice(0, 12);
    }

    // Parse transcript for message counts and costs
    try {
      const lines = fs.readFileSync(path.join(SESSIONS_DIR, file), 'utf8').split('\n');
      for (const line of lines) {
        if (!line) continue;
        try {
          const entry = JSON.parse(line);
          if (entry.type === 'message' && entry.message) {
            s.messageCount++;
            const cost = entry.message?.usage?.cost?.total;
            if (cost) {
              s.totalCost += cost;
              if (entry.message.timestamp > todayMs) {
                s.todayCost += cost;
              }
            }
          }
        } catch {}
      }
    } catch {}

    sessions.push(s);
  }

  sessions.sort((a, b) => b.updatedAt - a.updatedAt);
  return sessions;
}

function getHourlyActivity() {
  const now = Date.now();
  const cutoff = now - 24 * 60 * 60 * 1000;
  const buckets = Array.from({ length: 24 }, (_, i) => {
    const t = new Date(cutoff + (i + 1) * 3600000);
    return { hour: t.toTimeString().slice(0, 5), messages: 0, cost: 0 };
  });

  const files = fs.readdirSync(SESSIONS_DIR).filter(f => f.endsWith('.jsonl'));
  for (const file of files) {
    try {
      const lines = fs.readFileSync(path.join(SESSIONS_DIR, file), 'utf8').split('\n');
      for (const line of lines) {
        if (!line) continue;
        try {
          const entry = JSON.parse(line);
          if (entry.type !== 'message' || !entry.message?.timestamp) continue;
          const ts = entry.message.timestamp;
          if (ts < cutoff || ts > now) continue;
          const idx = Math.min(Math.floor((ts - cutoff) / 3600000), 23);
          buckets[idx].messages++;
          const cost = entry.message?.usage?.cost?.total;
          if (cost) buckets[idx].cost += cost;
        } catch {}
      }
    } catch {}
  }
  return buckets;
}

const server = http.createServer((req, res) => {
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Content-Type', 'application/json');

  if (req.url === '/api/dashboard') {
    try {
      const sessions = loadSessions();
      let totalCost = 0, todayCost = 0;
      for (const s of sessions) { totalCost += s.totalCost; todayCost += s.todayCost; }
      res.end(JSON.stringify({ sessions, totalCount: sessions.length, totalCost, todayCost }));
    } catch (e) {
      console.error('Dashboard error:', e);
      res.statusCode = 500;
      res.end(JSON.stringify({ error: e.message }));
    }
  } else if (req.url === '/api/hourly') {
    res.end(JSON.stringify(getHourlyActivity()));
  } else {
    res.statusCode = 404;
    res.end('{}');
  }
});

server.listen(PORT, () => {
  console.log(`Antenna dev API running on http://localhost:${PORT}`);
  console.log(`Reading sessions from ${SESSIONS_DIR}`);
});
