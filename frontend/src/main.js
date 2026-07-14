import './style.css';
import {
  ApplicationLists, AppVersion, CancelSync, DownloadApplication, Health, LastSync, LaunchApplication,
  RepairApplication, UninstallApplication, UpdateApplication,
} from '../wailsjs/go/main/App';
import { BrowserOpenURL, EventsOn } from '../wailsjs/runtime/runtime';

function escapeHTML(value) {
  return String(value)
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#039;');
}

function localRow(app) {
  const details = app.application;
  const updateAvailable = app.updateAvailable;
  const version = updateAvailable
    ? `${escapeHTML(details.version)} <span class="version-arrow">→</span> ${escapeHTML(app.availableVersion)} <span class="tag">可更新</span>`
    : escapeHTML(details.version);

  return `<tr>
    <td data-label="應用名"><strong>${escapeHTML(details.name)}</strong></td>
    <td data-label="版本號"><span class="version">${version}</span></td>
    <td data-label="作者">${escapeHTML(details.author)}</td>
    <td data-label="操作"><div class="actions">
      <button class="button primary launch-button" type="button" data-app-id="${escapeHTML(details.id)}" data-executable="${escapeHTML(details.executable)}">啟動</button>
      <button class="button secondary operation-button" data-operation="update" data-app-id="${escapeHTML(details.id)}" type="button" ${updateAvailable ? '' : 'disabled'}>更新</button>
      <button class="icon-button more-button" type="button" aria-label="更多操作" aria-haspopup="menu" data-app-id="${escapeHTML(details.id)}" data-app-name="${escapeHTML(details.name)}">…</button>
    </div></td>
  </tr>`;
}

async function launchApplication(button) {
  button.disabled = true;
  try {
    await LaunchApplication(button.dataset.appId, button.dataset.executable);
    window.setTimeout(() => { button.disabled = false; }, 3000);
  } catch (error) {
    button.disabled = false;
    window.alert(`無法啟動應用：${error}`);
  }
}

function intranetRow(app) {
  return `<tr>
    <td data-label="應用名"><strong>${escapeHTML(app.name)}</strong></td>
    <td data-label="版本號">${escapeHTML(app.version)}</td>
    <td data-label="作者">${escapeHTML(app.author)}</td>
    <td data-label="說明">${escapeHTML(app.description)}</td>
    <td data-label="操作"><button class="button download operation-button" data-operation="download" data-app-id="${escapeHTML(app.id)}" type="button"><span aria-hidden="true">↓</span> 下載</button></td>
  </tr>`;
}

document.querySelector('#app').innerHTML = `
  <main class="shell">
    <section class="panel local" aria-labelledby="local-title">
      <div class="panel-heading">
        <h2 id="local-title">本機可執行應用</h2>
      </div>
      <div class="table-wrap"><table>
        <thead><tr><th>應用名</th><th>版本號</th><th>作者</th><th>操作</th></tr></thead>
        <tbody id="local-apps"></tbody>
      </table></div>
    </section>

    <section class="panel intranet" aria-labelledby="intranet-title">
      <div class="panel-heading">
        <h2 id="intranet-title">內網可下載應用</h2>
      </div>
      <div class="table-wrap"><table>
        <thead><tr><th>應用名</th><th>版本號</th><th>作者</th><th>說明</th><th>操作</th></tr></thead>
        <tbody id="intranet-apps"></tbody>
      </table></div>
    </section>
    <footer class="statusbar" aria-label="系統狀態">
      <span id="operation-status" class="operation-status" role="status" aria-live="polite" hidden>
        <span id="operation-message"></span>
        <button id="cancel-sync" type="button">取消</button>
      </span>
      <span class="status"><span class="status-dot"></span>系統狀態：<strong id="system-status">啟動中</strong></span>
      <span class="separator"></span>
      <span>最後同步：<strong id="last-sync">尚未同步</strong></span>
      <button class="text-button" type="button">設定</button>
      <button id="open-about" class="text-button" type="button">關於</button>
    </footer>
    <div id="notice" class="notice" role="status" aria-live="polite" hidden></div>
    <div id="more-menu" class="more-menu" role="menu" hidden>
      <button class="operation-button" data-operation="repair" role="menuitem" type="button">修復</button>
      <button class="operation-button danger" data-operation="uninstall" role="menuitem" type="button">解除安裝</button>
    </div>
    <dialog id="about-dialog" class="about-dialog" aria-labelledby="about-title">
      <form method="dialog">
        <div class="about-heading">
          <h2 id="about-title">intra-launch</h2>
          <button class="about-close" value="close" aria-label="關閉" type="submit">×</button>
        </div>
        <p>公司內網使用的 portable 應用下載、更新與啟動工具。</p>
        <dl>
          <dt>版本</dt>
          <dd id="about-version"></dd>
          <dt>GitHub</dt>
          <dd><button id="open-github" class="link-button" type="button">github.com/honeymagico/intra-launch</button></dd>
          <dt>授權</dt>
          <dd>MIT License</dd>
        </dl>
        <div class="about-actions"><button class="button primary" value="close" type="submit">關閉</button></div>
      </form>
    </dialog>
  </main>`;

const aboutDialog = document.querySelector('#about-dialog');
document.querySelector('#open-about').addEventListener('click', async () => {
  document.querySelector('#about-version').textContent = await AppVersion();
  aboutDialog.showModal();
});
document.querySelector('#open-github').addEventListener('click', () => BrowserOpenURL('https://github.com/honeymagico/intra-launch'));
aboutDialog.addEventListener('click', (event) => {
  if (event.target === aboutDialog) aboutDialog.close();
});

document.querySelector('#local-apps').addEventListener('click', (event) => {
  const button = event.target.closest('.launch-button');
  if (button && !button.disabled) {
    launchApplication(button);
  }
  const operation = event.target.closest('.operation-button');
  if (operation && !operation.disabled) runOperation(operation);
});

document.querySelector('#intranet-apps').addEventListener('click', (event) => {
  const operation = event.target.closest('.operation-button');
  if (operation && !operation.disabled) runOperation(operation);
});

const moreMenu = document.querySelector('#more-menu');

document.addEventListener('click', (event) => {
  const moreButton = event.target.closest('.more-button');
  if (moreButton) {
    const rect = moreButton.getBoundingClientRect();
    moreMenu.querySelectorAll('.operation-button').forEach((button) => {
      button.dataset.appId = moreButton.dataset.appId;
      button.dataset.appName = moreButton.dataset.appName;
    });
    moreMenu.hidden = false;
    moreMenu.style.top = `${rect.bottom + 6}px`;
    moreMenu.style.left = `${Math.max(8, rect.right - moreMenu.offsetWidth)}px`;
    return;
  }

  const operation = event.target.closest('#more-menu .operation-button');
  if (operation) runOperation(operation);
  moreMenu.hidden = true;
});

const operations = {
  download: { label: '下載', call: DownloadApplication },
  update: { label: '更新', call: UpdateApplication },
  repair: { label: '修復', call: RepairApplication },
  uninstall: { label: '解除安裝', call: UninstallApplication },
};

const syncProgress = new Map();

function elapsedTime(startedAt) {
  const totalSeconds = Math.floor((Date.now() - startedAt) / 1000);
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = String(totalSeconds % 60).padStart(2, '0');
  return `${minutes}:${seconds}`;
}

function renderSyncProgress(state) {
  const status = document.querySelector('#operation-status');
  status.hidden = false;
  document.querySelector('#cancel-sync').hidden = false;
  document.querySelector('#operation-message').textContent = state.cancelRequested
    ? `${state.label}正在取消…`
    : `${state.label}中｜已完成 ${state.completedFiles} 個檔案｜經過 ${elapsedTime(state.startedAt)}`;
}

document.querySelector('#cancel-sync').addEventListener('click', async (event) => {
  const state = syncProgress.values().next().value;
  if (!state || state.cancelRequested) return;
  state.cancelRequested = true;
  event.currentTarget.disabled = true;
  renderSyncProgress(state);
  await CancelSync();
});

EventsOn('sync:progress', (progress) => {
  const state = syncProgress.get(progress.appId);
  if (!state) return;
  state.completedFiles = progress.completedFiles;
  renderSyncProgress(state);
});

let noticeTimer;

function showNotice(message, error = false) {
  const notice = document.querySelector('#notice');
  window.clearTimeout(noticeTimer);
  notice.hidden = false;
  notice.classList.toggle('error', error);
  notice.textContent = message;
  if (!error) {
    noticeTimer = window.setTimeout(() => {
      notice.hidden = true;
      notice.textContent = '';
    }, 4000);
  }
}

async function runOperation(button) {
  const operation = operations[button.dataset.operation];
  if (!operation) return;
  if (button.dataset.operation === 'uninstall' && !window.confirm(`確定要解除安裝「${button.dataset.appName}」嗎？`)) return;
  const original = button.innerHTML;
  const tracksProgress = ['download', 'update', 'repair'].includes(button.dataset.operation);
  let progressState = null;
  button.disabled = true;
  if (tracksProgress) {
    const lockedButtons = Array.from(document.querySelectorAll('.operation-button, .more-button, .launch-button')).map((item) => ({
      item,
      wasDisabled: item.disabled,
    }));
    lockedButtons.forEach(({ item }) => { item.disabled = true; });
    progressState = {
      button,
      label: operation.label,
      completedFiles: 0,
      startedAt: Date.now(),
      lockedButtons,
      cancelRequested: false,
    };
    progressState.timer = window.setInterval(() => renderSyncProgress(progressState), 1000);
    syncProgress.set(button.dataset.appId, progressState);
    document.querySelector('#cancel-sync').disabled = false;
    renderSyncProgress(progressState);
  } else {
    const operationStatus = document.querySelector('#operation-status');
    operationStatus.hidden = false;
    document.querySelector('#operation-message').textContent = `${operation.label}中…`;
    document.querySelector('#cancel-sync').hidden = true;
  }
  try {
    await operation.call(button.dataset.appId);
    showNotice(`${operation.label}完成。`);
    await refreshLists();
  } catch (error) {
    if (progressState?.cancelRequested) {
      showNotice(`${operation.label}已取消。`);
    } else {
      showNotice(`${operation.label}失敗：${error}`, true);
    }
  } finally {
    const state = syncProgress.get(button.dataset.appId);
    if (state?.button === button) {
      window.clearInterval(state.timer);
      syncProgress.delete(button.dataset.appId);
      state.lockedButtons.forEach(({ item, wasDisabled }) => {
        if (item.isConnected) item.disabled = wasDisabled;
      });
      const operationStatus = document.querySelector('#operation-status');
      operationStatus.hidden = true;
      document.querySelector('#operation-message').textContent = '';
    }
    if (!tracksProgress) {
      const operationStatus = document.querySelector('#operation-status');
      operationStatus.hidden = true;
      document.querySelector('#operation-message').textContent = '';
      document.querySelector('#cancel-sync').hidden = false;
    }
    if (button.isConnected) {
      button.innerHTML = original;
      button.disabled = false;
    }
  }
}

function emptyRow(columns, message) {
  return `<tr><td class="empty" colspan="${columns}">${escapeHTML(message)}</td></tr>`;
}

async function refreshLists() {
  const status = document.querySelector('#system-status');
  const lists = await ApplicationLists();
  document.querySelector('#local-apps').innerHTML = lists.installed.length ? lists.installed.map(localRow).join('') : emptyRow(4, '尚未安裝任何應用');
  document.querySelector('#intranet-apps').innerHTML = lists.available.length ? lists.available.map(intranetRow).join('') : emptyRow(5, lists.catalogError ? '目前無法取得內網應用清單' : '沒有可下載的應用');
  if (lists.catalogError) {
    status.textContent = lists.catalogError.includes('格式錯誤') || lists.catalogError.includes('內容錯誤') ? 'Catalog 錯誤' : 'SMB 無法連線';
    showNotice(lists.catalogError, true);
  } else {
    status.textContent = '正常';
    const lastSync = await LastSync();
    if (lastSync) document.querySelector('#last-sync').textContent = new Date(lastSync).toLocaleString('zh-TW');
  }
  if (lists.warnings?.length) showNotice(lists.warnings.join('；'), true);
}

async function verifyBackend() {
  const status = document.querySelector('#system-status');
  try {
    if (await Health() !== 'ready') {
      status.textContent = '異常';
      return;
    }
    await refreshLists();
  } catch {
    // Vite's browser-only preview has no Wails runtime; keep the UI preview usable.
    status.textContent = '介面預覽';
  }
}

verifyBackend();
