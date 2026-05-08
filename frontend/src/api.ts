import type { Category, SeedConfig, Site } from './types';

const BASE_URL = '/api';
// 核心：读取 .env 配置文件中的状态
const USE_MOCK = import.meta.env.VITE_USE_MOCK === 'true';

const getFaviconUrl = (domain: string) => `https://www.google.com/s2/favicons?domain=${domain}&sz=64`;

// --- 将原本在 App.vue 里的 Mock 数据转移到这里，保持 UI 层纯净 ---
const mockCategories: Category[] = [
  { id: '1', name: '电商', iconName: '🛒', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  { id: '2', name: '社交', iconName: '💬', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  { id: '3', name: '视频', iconName: '🎬', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
];

const mockSites: Site[] = [
  { id: '1', name: '淘宝', url: 'https://www.taobao.com', description: '亚洲较大的网上交易平台', categoryId: '1', icon: getFaviconUrl('www.taobao.com'), isPublic: true, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  { id: '2', name: '京东', url: 'https://www.jd.com', description: '专业的综合网上购物商城', categoryId: '1', icon: getFaviconUrl('www.jd.com'), isPublic: true, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  { id: '3', name: '微信', url: 'https://weixin.qq.com', description: '跨平台的通讯工具', categoryId: '2', icon: getFaviconUrl('weixin.qq.com'), isPublic: true, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
  { id: '4', name: 'B站', url: 'https://www.bilibili.com', description: '国内知名的视频弹幕网站', categoryId: '3', icon: getFaviconUrl('www.bilibili.com'), isPublic: true, createdAt: new Date().toISOString(), updatedAt: new Date().toISOString() },
];

// --- 真实网络请求封装 ---
async function fetchApi(endpoint: string, options: RequestInit = {}) {
  const token = localStorage.getItem('linkdock-token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { 'Authorization': `Bearer ${token}` } : {}),
    ...options.headers,
  };

  try {
    const response = await fetch(`${BASE_URL}${endpoint}`, { ...options, headers });
    if (!response.ok) {
      const errorText = await response.text().catch(() => '');
      let message = 'API 请求失败';
      try {
        const errorData = JSON.parse(errorText);
        message = errorData.message || message;
      } catch (e) {
        // 如果不是 JSON，尝试查看是否包含一些提示性的文本（比如 404 页面标题）
        if (errorText.includes('<title>')) {
          const title = errorText.match(/<title>(.*?)<\/title>/);
          if (title) message += `: ${title[1]}`;
        }
      }
      const error = new Error(message) as Error & { status?: number; statusText?: string; url?: string };
      error.status = response.status;
      error.statusText = response.statusText;
      error.url = response.url;
      throw error;
    }
    const data = await response.json();
    console.log(`API Success: ${endpoint}`, data);
    return data;
  } catch (err: any) {
    console.error(`API Error: ${endpoint}`, err);
    if (err.name === 'TypeError' && err.message === 'Failed to fetch') {
      throw new Error(`网络连接失败，请检查后端服务是否启动。 (URL: ${BASE_URL}${endpoint})`);
    }
    throw err;
  }
}

export const api = {
  // --- 认证相关 ---
  async login(credentials: any) {
    if (USE_MOCK) {
      if (credentials.username === 'admin' && credentials.password === '123456') {
        const user = { id: 'mock_admin', username: 'admin' };
        localStorage.setItem('linkdock-token', user.id);
        return user;
      }
      throw new Error('Mock: 账号或密码错误');
    }
    const data = await fetchApi('/auth/login', { method: 'POST', body: JSON.stringify(credentials) });
    localStorage.setItem('linkdock-token', data.id);
    return data;
  },

  async logout() {
    if (USE_MOCK) {
      localStorage.removeItem('linkdock-token');
      return;
    }
    await fetchApi('/auth/logout', { method: 'POST' });
    localStorage.removeItem('linkdock-token');
  },

  async getMe() {
    const token = localStorage.getItem('linkdock-token');
    if (!token) return null;
    if (USE_MOCK) return { id: 'mock_admin', username: 'admin' };
    try {
      const data = await fetchApi('/auth/me');
      if (!data?.authenticated) {
        localStorage.removeItem('linkdock-token');
        return null;
      }
      return data;
    } catch (e) {
      localStorage.removeItem('linkdock-token');
      return null;
    }
  },

  async register(credentials: any) {
    if (USE_MOCK) {
      const user = { id: 'mock_user_' + Date.now(), username: credentials.username };
      localStorage.setItem('linkdock-token', user.id);
      return user;
    }
    const data = await fetchApi('/auth/register', { method: 'POST', body: JSON.stringify(credentials) });
    localStorage.setItem('linkdock-token', data.id);
    return data;
  },
  async checkUsername(username: string) {
    if (USE_MOCK) {
      const reserved = ['admin', 'mock_admin'];
      return {
        available: !reserved.includes(username.toLowerCase()),
        message: reserved.includes(username.toLowerCase()) ? '该用户名已被占用' : '用户名可使用'
      };
    }
    return fetchApi('/auth/check-username', { method: 'POST', body: JSON.stringify({ username }) });
  },
  async getRecoveryQuestion(username: string) {
    if (USE_MOCK) {
      return { securityQuestion: '你最喜欢的网站是？' };
    }
    return fetchApi('/auth/recovery-question', { method: 'POST', body: JSON.stringify({ username }) });
  },
  async resetPassword(payload: { username: string; securityAnswer: string; newPassword: string }) {
    if (USE_MOCK) return { success: true };
    return fetchApi('/auth/reset-password', { method: 'POST', body: JSON.stringify(payload) });
  },
  async changePassword(payload: { oldPassword: string; newPassword: string }) {
    if (USE_MOCK) return { success: true };
    return fetchApi('/auth/change-password', { method: 'POST', body: JSON.stringify(payload) });
  },

  async getSeedData() {
    if (USE_MOCK) {
      return { success: true, categories: mockCategories, sites: mockSites };
    }
    return fetchApi('/admin/seed');
  },

  async saveSeedData(payload: SeedConfig) {
    if (USE_MOCK) return { success: true };
    return fetchApi('/admin/seed', { method: 'PUT', body: JSON.stringify(payload) });
  },

  // --- 数据获取 ---
  async getPublicData() {
    if (USE_MOCK) return { categories: mockCategories, sites: mockSites };
    return fetchApi('/public/data');
  },

  async getUserData(userId?: string) {
    if (USE_MOCK) return { categories: mockCategories, sites: mockSites };

    // 如果有 userId，拼接到 URL 参数中
    const url = userId ? `/user/data?user_id=${userId}` : '/user/data';
    return fetchApi(url);
  },

  // --- 变更操作 (Mock 下直接返回 success，防止页面报错) ---
  async createCategory(category: Category) {
    if (USE_MOCK) return;
    return fetchApi('/categories', { method: 'POST', body: JSON.stringify(category) });
  },
  async updateCategory(id: string, data: Partial<Category>) {
    if (USE_MOCK) return;
    return fetchApi(`/categories/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async deleteCategory(id: string) {
    if (USE_MOCK) return;
    return fetchApi(`/categories/${id}`, { method: 'DELETE' });
  },
  async createSite(site: Site) {
    if (USE_MOCK) return;
    return fetchApi('/sites', { method: 'POST', body: JSON.stringify(site) });
  },
  async updateSite(id: string, data: Partial<Site>) {
    if (USE_MOCK) return;
    return fetchApi(`/sites/${id}`, { method: 'PUT', body: JSON.stringify(data) });
  },
  async deleteSite(id: string) {
    if (USE_MOCK) return;
    return fetchApi(`/sites/${id}`, { method: 'DELETE' });
  },
  async updateCategoriesOrder(ids: string[]) {
    if (USE_MOCK) return;
    return fetchApi('/categories/reorder', {
      method: 'POST',
      body: JSON.stringify({ ids })
    });
  },
  async updateSitesOrder(sites: Array<{ id: string; categoryId: string; sortOrder: number }>) {
    if (USE_MOCK) return;
    return fetchApi('/sites/reorder', {
      method: 'POST',
      body: JSON.stringify({ sites })
    });
  },
  async trackSiteVisit(id: string) {
    if (USE_MOCK) return { success: true };
    return fetchApi(`/sites/${id}/visit`, { method: 'POST' });
  },
  async updateSiteState(id: string, payload: { isFavorite?: boolean; workflowStatus?: string }) {
    if (USE_MOCK) return { success: true, ...payload };
    return fetchApi(`/sites/${id}/state`, {
      method: 'POST',
      body: JSON.stringify(payload)
    });
  },
  async fetchUrlMetadata(url: string) {
    return fetchApi('/sites/fetch-metadata', {
      method: 'POST',
      body: JSON.stringify({ url })
    });
  }
};
