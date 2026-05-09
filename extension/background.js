const QUICK_SAVE_STORAGE_KEY = 'linkdock-pending-save-context';

async function ensureSidePanelBehavior() {
  try {
    await chrome.sidePanel.setPanelBehavior({ openPanelOnActionClick: true });
  } catch (error) {
    console.warn('Failed to configure side panel behavior:', error);
  }
}

function createContextMenus() {
  chrome.contextMenus.removeAll(() => {
    chrome.contextMenus.create({
      id: 'linkdock-save-page',
      title: '保存当前页到轻舟',
      contexts: ['page'],
    });
    chrome.contextMenus.create({
      id: 'linkdock-save-link',
      title: '保存这个链接到轻舟',
      contexts: ['link'],
    });
    chrome.contextMenus.create({
      id: 'linkdock-save-selection',
      title: '保存当前页并附带选中文本',
      contexts: ['selection'],
    });
  });
}

async function setPendingQuickSave(payload) {
  await chrome.storage.local.set({
    [QUICK_SAVE_STORAGE_KEY]: {
      ...payload,
      timestamp: Date.now(),
    },
  });
}

async function openQuickSavePanel(payload, tab) {
  if (!tab?.windowId) return;
  await setPendingQuickSave(payload);
  await ensureSidePanelBehavior();
  try {
    await chrome.sidePanel.open({ windowId: tab.windowId });
  } catch (error) {
    console.warn('Failed to open side panel for quick save:', error);
  }
}

chrome.runtime.onInstalled.addListener(() => {
  ensureSidePanelBehavior();
  createContextMenus();
});

chrome.runtime.onStartup?.addListener(() => {
  ensureSidePanelBehavior();
  createContextMenus();
});

chrome.contextMenus.onClicked.addListener(async (info, tab) => {
  if (!tab) return;

  if (info.menuItemId === 'linkdock-save-page') {
    await openQuickSavePanel({
      targetUrl: tab.url || '',
      targetName: tab.title || '',
    }, tab);
    return;
  }

  if (info.menuItemId === 'linkdock-save-link') {
    await openQuickSavePanel({
      targetUrl: info.linkUrl || tab.url || '',
      targetName: info.selectionText || info.linkText || tab.title || '',
      sourceUrl: tab.url || '',
    }, tab);
    return;
  }

  if (info.menuItemId === 'linkdock-save-selection') {
    await openQuickSavePanel({
      targetUrl: tab.url || '',
      targetName: tab.title || '',
      targetDescription: info.selectionText || '',
    }, tab);
  }
});

chrome.commands.onCommand.addListener(async (command) => {
  if (command !== 'quick-save-current-page') return;
  const [activeTab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!activeTab) return;
  await openQuickSavePanel({
    targetUrl: activeTab.url || '',
    targetName: activeTab.title || '',
  }, activeTab);
});
