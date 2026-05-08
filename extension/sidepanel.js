const APP_URL = window.LinkDockConfig.baseUrl;
const appFrame = document.getElementById('app-frame');
const QUICK_SAVE_STORAGE_KEY = 'linkdock-pending-save-context';
const QUICK_SAVE_MAX_AGE = 30 * 1000;

function buildPanelUrl(tab, pendingContext) {
  const tabUrl = tab?.url || '';
  const tabTitle = tab?.title || '';
  const params = new URLSearchParams({
    sidebar_mode: 'true',
    ext_url: tabUrl,
    ext_title: tabTitle,
  });

  if (pendingContext && Date.now() - pendingContext.timestamp < QUICK_SAVE_MAX_AGE) {
    params.set('quick_save', 'true');
    if (pendingContext.targetUrl) params.set('target_url', pendingContext.targetUrl);
    if (pendingContext.targetName) params.set('target_name', pendingContext.targetName);
    if (pendingContext.targetDescription) params.set('target_description', pendingContext.targetDescription);
  }

  return `${APP_URL}?${params.toString()}`;
}

async function getPendingSaveContext() {
  const result = await chrome.storage.local.get([QUICK_SAVE_STORAGE_KEY]);
  return result[QUICK_SAVE_STORAGE_KEY] || null;
}

async function clearExpiredPendingSaveContext() {
  const pendingContext = await getPendingSaveContext();
  if (!pendingContext) return null;
  if (Date.now() - pendingContext.timestamp > QUICK_SAVE_MAX_AGE) {
    await chrome.storage.local.remove(QUICK_SAVE_STORAGE_KEY);
    return null;
  }
  return pendingContext;
}

async function consumePendingSaveContext() {
  const pendingContext = await clearExpiredPendingSaveContext();
  if (pendingContext) {
    await chrome.storage.local.remove(QUICK_SAVE_STORAGE_KEY);
  }
  return pendingContext;
}

async function syncActiveTab() {
  const [activeTab] = await chrome.tabs.query({ active: true, currentWindow: true });
  const pendingContext = await consumePendingSaveContext();
  appFrame.src = buildPanelUrl(activeTab, pendingContext);
}

chrome.tabs.onActivated.addListener(() => {
  syncActiveTab().catch(() => {});
});

chrome.tabs.onUpdated.addListener((tabId, changeInfo, tab) => {
  if (!tab.active) return;
  if (changeInfo.status === 'complete' || changeInfo.title || changeInfo.url) {
    syncActiveTab().catch(() => {});
  }
});

chrome.storage.onChanged.addListener((changes, areaName) => {
  if (areaName === 'local' && changes[QUICK_SAVE_STORAGE_KEY]) {
    syncActiveTab().catch(() => {});
  }
});

syncActiveTab().catch(() => {
  appFrame.src = `${APP_URL}?sidebar_mode=true`;
});
