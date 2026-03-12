// ===== 配置 =====
const API_BASE = window.API_BASE || (
  (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1')
    ? 'http://127.0.0.1:6872/api'
    : '/api'
);

// ===== 用户 ID（持久化到 localStorage）=====
function getOrCreateUserId() {
  let uid = localStorage.getItem('oncall_user_id');
  if (!uid) {
    uid = 'user_' + Date.now().toString(36) + Math.random().toString(36).slice(2, 8);
    localStorage.setItem('oncall_user_id', uid);
  }
  return uid;
}

const userId = getOrCreateUserId();

// ===== 状态 =====
const state = {
  mode: 'qa',                  // 'qa' | 'aiops'
  sessionId: genId(),
  loading: false,
  pendingFile: null,           // { file, name, uploading, filePath }
  sessions: [],
  sidebarOpen: true,
};

// ===== 工具函数 =====
function genId() {
  return 'sess_' + Date.now().toString(36) + Math.random().toString(36).slice(2, 6);
}

function autoResize(el) {
  el.style.height = 'auto';
  el.style.height = Math.min(el.scrollHeight, 180) + 'px';
}

function esc(str) {
  if (typeof str !== 'string') return '';
  return str
    .replace(/&/g, '&amp;').replace(/</g, '&lt;')
    .replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function md(text) {
  if (!text) return '';

  // 1. 代码块（保护内容，不再被后续规则处理）
  const codeBlocks = [];
  text = text.replace(/```(\w*)\n?([\s\S]*?)```/g, (_, lang, code) => {
    const placeholder = `\x00CODE${codeBlocks.length}\x00`;
    codeBlocks.push(`<pre><code class="lang-${lang}">${esc(code.trim())}</code></pre>`);
    return placeholder;
  });

  // 2. 行内代码
  text = text.replace(/`([^`]+)`/g, (_, c) => `<code>${esc(c)}</code>`);

  // 3. 粗体 / 斜体
  text = text.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
  text = text.replace(/\*([^*]+)\*/g, '<em>$1</em>');

  // 4. 标题
  text = text.replace(/^### (.+)$/gm, '<h3>$1</h3>');
  text = text.replace(/^## (.+)$/gm, '<h2>$1</h2>');
  text = text.replace(/^# (.+)$/gm, '<h1>$1</h1>');

  // 5. 分割线
  text = text.replace(/^---$/gm, '<hr>');

  // 6. 无序列表（连续的 - / * 行合并为 <ul>）
  text = text.replace(/((?:^[ \t]*[-*][ \t]+.+\n?)+)/gm, match => {
    const items = match.trim().split('\n').map(line =>
      `<li>${line.replace(/^[ \t]*[-*][ \t]+/, '').trim()}</li>`
    ).join('');
    return `<ul>${items}</ul>`;
  });

  // 7. 有序列表
  text = text.replace(/((?:^[ \t]*\d+\.[ \t]+.+\n?)+)/gm, match => {
    const items = match.trim().split('\n').map(line =>
      `<li>${line.replace(/^[ \t]*\d+\.[ \t]+/, '').trim()}</li>`
    ).join('');
    return `<ol>${items}</ol>`;
  });

  // 8. 段落换行：非块级标签之间的换行转为 <br>
  text = text.replace(/(?<!>)\n(?!<(?:ul|ol|li|h[1-6]|hr|pre|blockquote))/g, '<br>');

  // 9. 还原代码块
  codeBlocks.forEach((block, i) => {
    text = text.replace(`\x00CODE${i}\x00`, block);
  });

  return text;
}

function scrollBottom() {
  const el = document.getElementById('messages');
  el.scrollTop = el.scrollHeight;
}

function setSendDisabled(v) {
  document.getElementById('sendBtn').disabled = v;
}

// ===== 欢迎屏 =====
function removeWelcome() {
  const w = document.getElementById('welcome');
  if (w) w.remove();
}

function fillInput(text) {
  const el = document.getElementById('inputEl');
  el.value = text;
  autoResize(el);
  el.focus();
}

// ===== 侧边栏 =====
function toggleSidebar() {
  state.sidebarOpen = !state.sidebarOpen;
  const sidebar = document.getElementById('sidebar');
  const openBtn  = document.getElementById('sidebarOpenBtn');
  sidebar.classList.toggle('collapsed', !state.sidebarOpen);
  openBtn.style.display = state.sidebarOpen ? 'none' : 'flex';
}

// ===== 模式切换 =====
function selectMode(mode) {
  state.mode = mode;
  const isQa = mode === 'qa';

  // 侧边栏导航高亮
  document.getElementById('navQa').classList.toggle('active', isQa);
  document.getElementById('navAiops').classList.toggle('active', !isQa);

  // 输入框 pill
  document.getElementById('modePillLabel').textContent = isQa ? '业务问答' : 'AI Ops';
  document.getElementById('modeOptQa').classList.toggle('active', isQa);
  document.getElementById('modeOptAiops').classList.toggle('active', !isQa);

  // 顶部栏
  document.getElementById('topbarTitle').textContent = isQa ? '业务问答' : 'AI Ops';
  document.getElementById('topbarBadge').textContent = isQa ? '流式输出' : '流式推送';

  closeModeMenu();
}

function toggleModeMenu() {
  const menu = document.getElementById('modeDropdown');
  menu.style.display = menu.style.display === 'none' ? 'block' : 'none';
}

function closeModeMenu() {
  document.getElementById('modeDropdown').style.display = 'none';
}

// 点击外部关闭
document.addEventListener('click', e => {
  const pill = document.getElementById('modePill');
  if (!pill.contains(e.target)) closeModeMenu();
});

// ===== 新建对话 =====
function newChat() {
  state.sessionId = genId();
  state.loading = false;
  state.pendingFile = null;

  const msgs = document.getElementById('messages');
  msgs.innerHTML = `
    <div class="welcome" id="welcome">
      <div class="welcome-icon">
        <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path d="M12 2L2 7l10 5 10-5-10-5z"/>
          <path d="M2 17l10 5 10-5"/>
          <path d="M2 12l10 5 10-5"/>
        </svg>
      </div>
      <h1 class="welcome-title">你好，我是智能 OnCall 助手</h1>
      <p class="welcome-sub">选择左侧功能模式，开始提问</p>
      <div class="welcome-cards">
        <div class="welcome-card" onclick="fillInput('最近有哪些告警？')">
          <div class="card-icon">🔔</div>
          <div class="card-text">最近有哪些告警？</div>
        </div>
        <div class="welcome-card" onclick="fillInput('帮我分析一下服务异常原因')">
          <div class="card-icon">🔍</div>
          <div class="card-text">帮我分析服务异常原因</div>
        </div>
        <div class="welcome-card" onclick="fillInput('CPU 使用率过高如何排查？')">
          <div class="card-icon">⚡</div>
          <div class="card-text">CPU 使用率过高如何排查？</div>
        </div>
      </div>
    </div>
  `;

  const input = document.getElementById('inputEl');
  input.value = '';
  input.style.height = 'auto';
  clearFile();
  renderHistory();
  setSendDisabled(false);
  input.focus();
}

// ===== 会话历史（localStorage 管理）=====

const SESSIONS_KEY = 'oncall_sessions';

// loadSessions 从 localStorage 读取当前用户的会话列表
function loadSessions() {
  try {
    const raw = localStorage.getItem(SESSIONS_KEY);
    const list = raw ? JSON.parse(raw) : [];
    // 倒序展示（最新在前）
    state.sessions = list.sort((a, b) => b.createdAt - a.createdAt);
  } catch (e) {
    state.sessions = [];
  }
  renderHistory();
}

// _persistSessions 将 state.sessions 写回 localStorage
function _persistSessions() {
  try {
    localStorage.setItem(SESSIONS_KEY, JSON.stringify(state.sessions));
  } catch (e) {
    console.warn('保存会话历史失败:', e);
  }
}

// saveSession 首次发送消息时将当前会话记录到本地列表，title 为用户第一个问题
function saveSession(title) {
  if (state.sessions.find(s => s.id === state.sessionId)) return;
  const session = {
    id: state.sessionId,
    title: title.slice(0, 26) + (title.length > 26 ? '…' : ''),
    createdAt: Date.now(),
  };
  state.sessions.unshift(session);
  _persistSessions();
  renderHistory();
}

function renderHistory() {
  const list = document.getElementById('historyList');
  if (!state.sessions.length) {
    list.innerHTML = '<div class="history-empty">暂无对话记录</div>';
    return;
  }
  list.innerHTML = state.sessions.map(s => `
    <div class="history-item ${s.id === state.sessionId ? 'active' : ''}" onclick="switchSession('${s.id}')">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
      </svg>
      <span class="history-item-text">${esc(s.title)}</span>
      <button class="history-item-del" title="删除会话" onclick="event.stopPropagation(); deleteSession('${s.id}')">
        <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <polyline points="3 6 5 6 21 6"/>
          <path d="M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/>
          <path d="M10 11v6M14 11v6"/>
          <path d="M9 6V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
        </svg>
      </button>
    </div>
  `).join('');
}

function switchSession(id) {
  if (state.sessionId === id) return;

  // 保存当前会话消息
  saveMessages(state.sessionId);

  state.sessionId = id;
  renderHistory();

  // 恢复目标会话消息
  const restored = loadMessages(id);
  if (!restored) {
    // 该会话无消息记录，显示欢迎屏
    const msgs = document.getElementById('messages');
    msgs.innerHTML = `
      <div class="welcome" id="welcome">
        <div class="welcome-icon">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
        </div>
        <h1 class="welcome-title">你好，我是智能 OnCall 助手</h1>
        <p class="welcome-sub">选择左侧功能模式，开始提问</p>
        <div class="welcome-cards">
          <div class="welcome-card" onclick="fillInput('最近有哪些告警？')">
            <div class="card-icon">🔔</div>
            <div class="card-text">最近有哪些告警？</div>
          </div>
          <div class="welcome-card" onclick="fillInput('帮我分析一下服务异常原因')">
            <div class="card-icon">🔍</div>
            <div class="card-text">帮我分析服务异常原因</div>
          </div>
          <div class="welcome-card" onclick="fillInput('CPU 使用率过高如何排查？')">
            <div class="card-icon">⚡</div>
            <div class="card-text">CPU 使用率过高如何排查？</div>
          </div>
        </div>
      </div>
    `;
  }
}

// deleteSession 从 localStorage 删除指定会话及其消息记录
function deleteSession(sessionId) {
  state.sessions = state.sessions.filter(s => s.id !== sessionId);
  _persistSessions();
  deleteMessages(sessionId);

  // 若删除的是当前会话，则新建一个
  if (state.sessionId === sessionId) {
    newChat();
  } else {
    renderHistory();
  }
  showToast('会话已删除', 'info');
}

// ===== 会话消息持久化 =====

// _msgKey 返回指定会话的消息存储 key
function _msgKey(sessionId) {
  return 'oncall_msgs_' + sessionId;
}

// saveMessages 将当前会话的消息 HTML 持久化到 localStorage
function saveMessages(sessionId) {
  const msgs = document.getElementById('messages');
  // 不保存欢迎屏
  if (msgs.querySelector('#welcome')) return;
  try {
    localStorage.setItem(_msgKey(sessionId), msgs.innerHTML);
  } catch (e) {
    console.warn('保存消息失败:', e);
  }
}

// loadMessages 从 localStorage 恢复指定会话的消息，返回是否有内容
function loadMessages(sessionId) {
  try {
    const html = localStorage.getItem(_msgKey(sessionId));
    if (html) {
      document.getElementById('messages').innerHTML = html;
      scrollBottom();
      return true;
    }
  } catch (e) {
    console.warn('读取消息失败:', e);
  }
  return false;
}

// deleteMessages 删除指定会话的消息记录
function deleteMessages(sessionId) {
  try {
    localStorage.removeItem(_msgKey(sessionId));
  } catch (e) {}
}

// ===== 键盘快捷键 =====
document.getElementById('inputEl').addEventListener('keydown', e => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
});

// ===== 文件上传 =====
function triggerUpload() {
  document.getElementById('fileInput').click();
}

function handleFileSelect(input) {
  const file = input.files[0];
  if (!file) return;
  input.value = '';

  state.pendingFile = { file, name: file.name, uploading: true, filePath: null };
  showFileChip(file.name, true);
  doUpload(file);
}

async function doUpload(file) {
  const fd = new FormData();
  fd.append('file', file);

  try {
    const res = await fetch(`${API_BASE}/upload`, { method: 'POST', body: fd });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);

    const data = await res.json();
    const info = data?.data || data;
    const filePath = info.filePath || info.fileName || file.name;

    if (state.pendingFile?.name === file.name) {
      state.pendingFile.uploading = false;
      state.pendingFile.filePath = filePath;
      showFileChip(file.name, false);
    }
  } catch (err) {
    if (state.pendingFile?.name === file.name) clearFile();
    console.error('上传失败:', err);
    showToast(`文件上传失败: ${err.message}`, 'error');
  }
}

function showFileChip(name, uploading) {
  document.getElementById('fileChipName').textContent = name;
  document.getElementById('fileChipStatus').textContent = uploading ? '上传中…' : '';
  document.getElementById('fileChipWrap').style.display = 'block';
}

function clearFile() {
  state.pendingFile = null;
  document.getElementById('fileChipWrap').style.display = 'none';
  document.getElementById('fileChipName').textContent = '';
  document.getElementById('fileChipStatus').textContent = '';
}

function removeFile() { clearFile(); }

// ===== Toast 提示 =====
function showToast(msg, type = 'info') {
  const t = document.createElement('div');
  t.style.cssText = `
    position:fixed; bottom:80px; left:50%; transform:translateX(-50%);
    background:${type === 'error' ? '#ef4444' : '#10a37f'};
    color:#fff; padding:8px 18px; border-radius:8px;
    font-size:13px; z-index:9999; box-shadow:0 4px 16px rgba(0,0,0,.4);
    animation: fadeUp .2s ease;
  `;
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(() => t.remove(), 3000);
}

// ===== 消息渲染 =====
function appendMsg(role, htmlContent) {
  removeWelcome();
  const container = document.getElementById('messages');
  const wrap = document.createElement('div');
  wrap.className = 'msg-wrap';
  const isUser = role === 'user';
  wrap.innerHTML = `
    <div class="msg-row ${isUser ? 'user' : ''}">
      <div class="avatar ${isUser ? 'user' : 'ai'}">${isUser ? 'U' : 'AI'}</div>
      <div class="bubble ${isUser ? 'user' : 'ai'}">${htmlContent}</div>
    </div>
  `;
  container.appendChild(wrap);
  scrollBottom();
  return wrap.querySelector('.bubble');
}

// ===== 发送入口 =====
async function sendMessage() {
  if (state.loading) return;

  const input = document.getElementById('inputEl');
  const question = input.value.trim();
  const file = state.pendingFile;

  if (!question && !file) return;

  // 文件还在上传中
  if (file?.uploading) {
    showToast('文件仍在上传中，请稍候…', 'info');
    return;
  }

  // 清空输入
  input.value = '';
  input.style.height = 'auto';
  clearFile();

  // 保存会话
  saveSession(question || file?.name || '文件上传');

  // 用户气泡
  if (file) {
    appendMsg('user', `
      <div class="file-attach">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        ${esc(file.name)}
      </div>
      ${question ? `<div>${esc(question)}</div>` : ''}
    `);
  } else {
    appendMsg('user', esc(question));
  }

  if (!question) return;

  if (state.mode === 'qa') {
    await doStreamQA(question);
  } else {
    await doStreamAIOps(question);
  }
}

// ===== 业务问答：/chat_stream（SSE）=====
async function doStreamQA(question) {
  state.loading = true;
  setSendDisabled(true);

  // 创建 AI 气泡
  removeWelcome();
  const container = document.getElementById('messages');
  const wrap = document.createElement('div');
  wrap.className = 'msg-wrap';
  wrap.innerHTML = `
    <div class="msg-row">
      <div class="avatar ai">AI</div>
      <div class="bubble ai"><span class="cursor"></span></div>
    </div>
  `;
  container.appendChild(wrap);
  const bubble = wrap.querySelector('.bubble');
  scrollBottom();

  let fullText = '';

  try {
    const res = await fetch(`${API_BASE}/chat_stream`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ userId, sessionId: state.sessionId, question: question }),
    });

    if (!res.ok) {
      bubble.innerHTML = `<span style="color:var(--c-error)">请求失败 (${res.status})</span>`;
      return;
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value, { stream: true });
      for (const line of chunk.split('\n')) {
        if (!line.startsWith('data:')) continue;
        const raw = line.slice(5).trim();
        if (raw === '[DONE]') break;
        // 过滤结束标记
        if (raw === '[DONE]' || raw === 'Stream completed' || raw === '') continue;
        // 还原后端转义的换行符
        const content = raw.replace(/\\n/g, '\n');
        try {
          const parsed = JSON.parse(content);
          const text = parsed?.data || parsed?.content || parsed?.message || '';
          if (text && text !== '[DONE]' && text !== 'Stream completed') {
            fullText += text;
            bubble.innerHTML = md(fullText) + '<span class="cursor"></span>';
            scrollBottom();
          }
        } catch {
          // 非 JSON 纯文本内容
          fullText += content;
          bubble.innerHTML = md(fullText) + '<span class="cursor"></span>';
          scrollBottom();
        }
      }
    }

    bubble.innerHTML = md(fullText) || '<span style="color:var(--c-text-3)">（无内容）</span>';
    // 流结束后持久化消息
    saveMessages(state.sessionId);
  } catch (err) {
    bubble.innerHTML = `<span style="color:var(--c-error)">网络错误: ${esc(err.message)}</span>`;
  } finally {
    state.loading = false;
    setSendDisabled(false);
    document.getElementById('inputEl').focus();
  }
}

// ===== AI Ops：/ai_ops（SSE 流式）=====
async function doStreamAIOps(question) {
  state.loading = true;
  setSendDisabled(true);

  // 创建 AI 气泡
  removeWelcome();
  const container = document.getElementById('messages');
  const wrap = document.createElement('div');
  wrap.className = 'msg-wrap';
  wrap.innerHTML = `
    <div class="msg-row">
      <div class="avatar ai">AI</div>
      <div class="bubble ai">
        <div class="aiops-intent" id="aiopsIntent" style="display:none"></div>
        <div class="aiops-steps" id="aiopsSteps"></div>
        <div class="thinking" id="aiopsThinking"><span></span><span></span><span></span></div>
        <div class="aiops-result" id="aiopsResult" style="display:none"></div>
      </div>
    </div>
  `;
  container.appendChild(wrap);
  const intentEl  = wrap.querySelector('#aiopsIntent');
  const stepsEl   = wrap.querySelector('#aiopsSteps');
  const thinkEl   = wrap.querySelector('#aiopsThinking');
  const resultEl  = wrap.querySelector('#aiopsResult');
  scrollBottom();

  let stepCount = 0;

  try {
    const res = await fetch(`${API_BASE}/ai_ops`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ userId, sessionId: state.sessionId, query: question }),
    });

    if (!res.ok) {
      thinkEl.remove();
      resultEl.style.display = 'block';
      resultEl.innerHTML = `<span style="color:var(--c-error)">请求失败 (${res.status})</span>`;
      return;
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value, { stream: true });
      for (const line of chunk.split('\n')) {
        if (!line.startsWith('data:')) continue;
        const raw = line.slice(5).trim();
        if (!raw) continue;

        try {
          const parsed = JSON.parse(raw);
          const evtType = parsed?.event || parsed?.type || '';
          const evtData = parsed?.data || parsed?.content || '';

          if (evtType === 'intent') {
            // 意图确认事件：在步骤列表上方展示系统对用户需求的理解
            intentEl.style.display = 'block';
            intentEl.innerHTML = `
              <div class="intent-icon">🎯</div>
              <span>${esc(evtData)}</span>
            `;
            scrollBottom();

          } else if (evtType === 'step' || (!evtType && evtData && !parsed?.result)) {
            // 步骤事件：追加步骤卡片
            stepCount++;
            const stepDiv = document.createElement('div');
            stepDiv.className = 'aiops-step';
            stepDiv.innerHTML = `
              <div class="step-icon">
                <svg width="8" height="8" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
              <span>${md(evtData || raw)}</span>
            `;
            stepsEl.appendChild(stepDiv);
            scrollBottom();

          } else if (evtType === 'done') {
            // 完成事件：evtData 是 JSON 字符串 {scene, result}，解析后以报告样式展示
            thinkEl.remove();
            let finalResult = evtData;
            try {
              const donePayload = JSON.parse(evtData);
              finalResult = donePayload?.result || evtData;
            } catch { /* evtData 非 JSON，直接使用原文 */ }
            resultEl.style.display = 'block';
            resultEl.innerHTML = `
              <div class="report-header">
                <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" y1="13" x2="8" y2="13"/>
                  <line x1="16" y1="17" x2="8" y2="17"/>
                  <polyline points="10 9 9 9 8 9"/>
                </svg>
                分析报告
              </div>
              <div class="report-body">${md(finalResult)}</div>
            `;
            scrollBottom();

          } else if (evtType === 'error') {
            thinkEl.remove();
            resultEl.style.display = 'block';
            resultEl.innerHTML = `<span style="color:var(--c-error)">错误: ${esc(evtData)}</span>`;
            scrollBottom();
          }
        } catch {
          // 非 JSON，当作步骤文本
          if (raw && raw !== '[DONE]') {
            stepCount++;
            const stepDiv = document.createElement('div');
            stepDiv.className = 'aiops-step';
            stepDiv.innerHTML = `
              <div class="step-icon">
                <svg width="8" height="8" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
              <span>${esc(raw)}</span>
            `;
            stepsEl.appendChild(stepDiv);
            scrollBottom();
          }
        }
      }
    }

    // 流结束后若 thinking 还在，移除
    if (thinkEl.parentNode) thinkEl.remove();
    if (resultEl.style.display === 'none' && stepCount > 0) {
      resultEl.style.display = 'block';
      resultEl.innerHTML = '<span style="color:var(--c-text-3)">分析完成</span>';
    }
    // 流结束后持久化消息
    saveMessages(state.sessionId);

  } catch (err) {
    if (thinkEl.parentNode) thinkEl.remove();
    resultEl.style.display = 'block';
    resultEl.innerHTML = `<span style="color:var(--c-error)">网络错误: ${esc(err.message)}</span>`;
  } finally {
    state.loading = false;
    setSendDisabled(false);
    document.getElementById('inputEl').focus();
  }
}

// ===== 初始化 =====
loadSessions();