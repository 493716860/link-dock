<script setup lang="ts">
import {ref, computed, onMounted, watch} from 'vue';
import Sidebar from '@/components/Sidebar.vue';
import Header from '@/components/Header.vue';
import SiteCard from '@/components/SiteCard.vue';
import CategoryForm from '@/components/CategoryForm.vue';
import SiteForm from '@/components/SiteForm.vue';
import LoginModal from '@/components/LoginModal.vue';
import ChangePasswordModal from '@/components/ChangePasswordModal.vue';
import CommandPalette from '@/components/CommandPalette.vue';
import AdminSeedManager from '@/components/AdminSeedManager.vue';
import OfficeToolbox from '@/components/OfficeToolbox.vue';
import {api} from './api';
import type {Category, Site, WorkflowStatus} from './types';
import {Loader2, Search, SearchX, Plus, Star, Clock3, Inbox, History, Sparkles, ChevronDown, ChevronLeft, ChevronRight, Github, AlertTriangle} from 'lucide-vue-next';
import { onKeyStroke } from '@vueuse/core';
import { useRoute, useRouter } from 'vue-router';
import draggable from 'vuedraggable';

type SiteGroup = {
  category: Category;
  sites: Site[];
};

const currentUser = ref<{ id: string, username: string, isSuperAdmin?: boolean } | null>(null);
const route = useRoute();
const router = useRouter();
const activeCategoryId = ref<string | null>(null);
const searchQuery = ref('');
const quickFilter = ref<'all' | 'recent_added' | 'recent_visited' | 'favorites' | 'read_later' | 'unorganized'>('all');

const categories = ref<Category[]>([]);
const sites = ref<Site[]>([]);
const draggableSiteGroups = ref<SiteGroup[]>([]);
const draggingSiteId = ref<string | null>(null);
const isDroppingToSidebarCategory = ref(false);
const isDataLoaded = ref(false);
const isSidebarMode = ref(false);
const isSidebarCollapsed = ref(false);
const connectionIssue = ref(false);

const retryFetch = () => {
  window.location.reload();
}

const showCategoryForm = ref(false);
const editingCategory = ref<Category | null>(null);

const showSiteForm = ref(false);
const editingSite = ref<Site | null>(null);
const initialSiteData = ref<{ url?: string, name?: string, description?: string } | null>(null);
const currentPageData = ref<{ url?: string, name?: string, description?: string } | null>(null);
const preferredCategoryId = ref<string | null>(null);
const pendingQuickSave = ref(false);

const showLoginModal = ref(false);
const initialIsRegister = ref(false);

const showCommandPalette = ref(false);
const showChangePasswordModal = ref(false);
const showCompactCategoryPanel = ref(false);
const showCompactFilterPanel = ref(false);
const visibleSiteLimit = ref(60);
const duplicateSaveMessage = ref('');
const showOfficeToolbox = ref(false);

const RECENT_CATEGORY_KEY = 'linkdock-recent-category-id';
const OPEN_SOURCE_URL = 'https://github.com/493716860/link-dock';
const SITE_PAGE_SIZE = 60;

const normalizeUrl = (raw?: string) => {
  const target = (raw || '').trim();
  if (!target) return '';
  try {
    const parsed = new URL(target.startsWith('http://') || target.startsWith('https://') ? target : `https://${target}`);
    parsed.hash = '';
    if (parsed.pathname === '/') parsed.pathname = '';
    parsed.hostname = parsed.hostname.toLowerCase();
    return parsed.toString();
  } catch {
    return '';
  }
};

const getHostname = (raw?: string) => {
  const normalized = normalizeUrl(raw);
  if (!normalized) return '';
  try {
    return new URL(normalized).hostname;
  } catch {
    return '';
  }
};

const recentCategoryId = computed(() => window.localStorage.getItem(RECENT_CATEGORY_KEY));

const canEdit = computed(() => true);
const isCompactWorkspace = computed(() => isSidebarMode.value);
const isAdminSeedRoute = computed(() => route.name === 'admin-seed');
const canShowAdminSeedPage = computed(() => {
  return isAdminSeedRoute.value && !isCompactWorkspace.value && isDataLoaded.value && !!currentUser.value?.isSuperAdmin;
});

const openBookmarkHome = () => {
  showOfficeToolbox.value = false;
  router.push({ name: 'home' });
};

const openOfficeToolbox = () => {
  if (isAdminSeedRoute.value) {
    router.push({ name: 'home' });
  }
  showOfficeToolbox.value = true;
  showCompactCategoryPanel.value = false;
  showCompactFilterPanel.value = false;
};

const toggleSidebarCollapse = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value;
};

const currentPageExistingSite = computed(() => {
  if (!currentUser.value) return null;
  const normalized = normalizeUrl(currentPageData.value?.url);
  if (!normalized) return null;
  return sites.value.find(site => normalizeUrl(site.url) === normalized) || null;
});

const currentPageSameDomainSites = computed(() => {
  const hostname = getHostname(currentPageData.value?.url);
  if (!hostname) return [];
  return sites.value.filter(site => {
    if (currentPageExistingSite.value?.id === site.id) return false;
    return getHostname(site.url) === hostname;
  });
});

const recommendedCategoryIdForCurrentPage = computed(() => {
  const grouped = new Map<string, number>();
  const sourceSites = currentPageSameDomainSites.value;
  sourceSites.forEach(site => {
    grouped.set(site.categoryId, (grouped.get(site.categoryId) || 0) + 1);
  });
  if (grouped.size > 0) {
    return [...grouped.entries()].sort((a, b) => b[1] - a[1])[0][0];
  }
  return recentCategoryId.value || activeCategoryId.value || null;
});

const recommendedCategoryForCurrentPage = computed(() => {
  return categories.value.find(cat => String(cat.id) === String(recommendedCategoryIdForCurrentPage.value)) || null;
});

const quickFilters = [
  { id: 'all', label: '全部', icon: Search },
  { id: 'recent_added', label: '最近新增', icon: Sparkles },
  { id: 'recent_visited', label: '最近访问', icon: History },
  { id: 'favorites', label: '常用', icon: Star },
  { id: 'read_later', label: '稍后读', icon: Clock3 },
  { id: 'unorganized', label: '未整理', icon: Inbox },
] as const;

const activeQuickFilterLabel = computed(() => {
  return quickFilters.find(filter => filter.id === quickFilter.value)?.label || '全部';
});

onKeyStroke(['k', 'K'], (e) => {
  if ((e.ctrlKey || e.metaKey)) {
    e.preventDefault();
    showCommandPalette.value = true;
  }
});

const handleCommandSelectSite = async (site: Site) => {
  await handleOpenSite(site);
}
const handleCommandSelectCategory = (id: string | null) => {
  activeCategoryId.value = id;
}

const openLogin = (register = false) => {
  initialIsRegister.value = register;
  showLoginModal.value = true;
}

const openSaveCurrentPage = () => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  if (currentPageExistingSite.value) {
    const categoryName = categories.value.find(category => String(category.id) === String(currentPageExistingSite.value?.categoryId))?.name || '已有分类';
    duplicateSaveMessage.value = `这个网址已经保存过了，位于「${categoryName}」中。`;
    return;
  }
  duplicateSaveMessage.value = '';
  editingSite.value = null;
  initialSiteData.value = currentPageData.value;
  preferredCategoryId.value = recommendedCategoryIdForCurrentPage.value;
  showSiteForm.value = true;
}

const quickSaveCurrentPage = async () => {
  if (!currentPageData.value?.url) return;
  if (!currentUser.value) {
    openLogin();
    return;
  }
  if (currentPageExistingSite.value) {
    const categoryName = categories.value.find(category => String(category.id) === String(currentPageExistingSite.value?.categoryId))?.name || '已有分类';
    duplicateSaveMessage.value = `这个网址已经保存过了，位于「${categoryName}」中。`;
    pendingQuickSave.value = false;
    return;
  }

  const categoryId = recommendedCategoryIdForCurrentPage.value
      || recentCategoryId.value
      || activeCategoryId.value
      || categories.value[0]?.id;

  if (!categoryId) {
    openSaveCurrentPage();
    return;
  }

  let name = currentPageData.value.name || currentPageData.value.url;
  let description = currentPageData.value.description || '';
  let icon = '';

  try {
    const meta = await api.fetchUrlMetadata(currentPageData.value.url);
    if (meta?.success) {
      name = name || meta.title || currentPageData.value.url;
      description = description || meta.description || '';
      icon = meta.icon || '';
    }
  } catch (error) {
    console.warn('Quick save metadata fetch skipped:', error);
  }

  try {
    await api.createSite({
      name,
      url: currentPageData.value.url,
      description,
      categoryId,
      icon,
      tagsText: '',
      workflowStatus: 'unorganized',
      isFavorite: false,
      isPublic: true,
    } as Site);
    window.localStorage.setItem(RECENT_CATEGORY_KEY, categoryId);
    duplicateSaveMessage.value = '';
    pendingQuickSave.value = false;
    await fetchData();
  } catch (error: any) {
    if (error?.status === 409) {
      duplicateSaveMessage.value = '这个网址已经保存过了，不需要重复添加。';
      pendingQuickSave.value = false;
      await fetchData();
      return;
    }
    openSaveCurrentPage();
  }
}

const openChangePassword = () => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  showChangePasswordModal.value = true;
}

const openAdminSeedManager = () => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  if (!currentUser.value.isSuperAdmin) {
    alert('只有超级管理员可以维护默认书签');
    return;
  }
  activeCategoryId.value = null;
  searchQuery.value = '';
  quickFilter.value = 'all';
  router.push({ name: 'admin-seed' });
}

const MOCK_CATEGORIES: Category[] = [
  { id: '1', name: '常用工具', iconName: '🛠️' },
  { id: '2', name: '设计资源', iconName: '🎨' },
  { id: '3', name: '开发社区', iconName: '💻' },
  { id: '4', name: '摸鱼专区', iconName: '☕' },
];

const MOCK_SITES: Site[] = [
  { id: '1', categoryId: '1', name: 'Google', url: 'https://www.google.com', description: '全球最大的搜索引擎', icon: 'https://www.google.com/favicon.ico' },
  { id: '2', categoryId: '1', name: 'ChatGPT', url: 'https://chat.openai.com', description: '强大的 AI 助手', icon: 'https://chat.openai.com/favicon.ico' },
  { id: '3', categoryId: '2', name: 'Dribbble', url: 'https://dribbble.com', description: '寻找设计灵感', icon: 'https://dribbble.com/favicon.ico' },
  { id: '4', categoryId: '2', name: 'Behance', url: 'https://www.behance.net', description: '展示和发现创意作品', icon: 'https://www.behance.net/favicon.ico' },
  { id: '5', categoryId: '3', name: 'GitHub', url: 'https://github.com', description: '面向开源及私有软件项目的托管平台', icon: 'https://github.com/favicon.ico' },
  { id: '6', categoryId: '3', name: 'V2EX', url: 'https://v2ex.com', description: '创意工作者们的社区', icon: 'https://v2ex.com/favicon.ico' }
];

const buildSiteGroups = (sourceSites: Site[]) => {
  const groups: SiteGroup[] = categories.value.map(cat => ({
    category: cat,
    sites: sourceSites
        .filter(s => String(s.categoryId) === String(cat.id))
        .sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0))
  }));

  const categorizedSiteIds = new Set<string>();
  groups.forEach(group => {
    group.sites.forEach(site => categorizedSiteIds.add(site.id));
  });

  const orphanSites = sourceSites.filter(s => !categorizedSiteIds.has(s.id));
  if (orphanSites.length > 0) {
    groups.push({
      category: { id: 'orphan', name: '未分类', iconName: '📁' } as Category,
      sites: [...orphanSites].sort((a, b) => (a.sortOrder ?? 0) - (b.sortOrder ?? 0))
    });
  }

  return groups;
};

const syncActiveCategoryWithData = (nextCategories: Category[]) => {
  if (!activeCategoryId.value) {
    return;
  }

  const exists = nextCategories.some(category => String(category.id) === String(activeCategoryId.value));
  if (!exists) {
    activeCategoryId.value = null;
  }
};

const syncDraggableSiteGroups = () => {
  draggableSiteGroups.value = buildSiteGroups(sites.value);
};

const cycleWorkflowStatusValue = (current?: WorkflowStatus) => {
  if (current === 'unorganized') return 'read_later';
  if (current === 'read_later') return 'none';
  return 'unorganized';
};

const fetchData = async () => {
  try {
    const data = currentUser.value
        ? await api.getUserData(currentUser.value.id)
        : await api.getPublicData();
    if (!currentUser.value) {
      quickFilter.value = 'all';
    }
    categories.value = data.categories && data.categories.length > 0 ? data.categories : MOCK_CATEGORIES;
    sites.value = data.sites && data.sites.length > 0 ? data.sites : MOCK_SITES;
    syncActiveCategoryWithData(categories.value);
    syncDraggableSiteGroups();
  } catch (err: any) {
    console.error("Fetch data error:", err);
    if (!currentUser.value) {
      quickFilter.value = 'all';
    }
    categories.value = MOCK_CATEGORIES;
    sites.value = MOCK_SITES;
    syncActiveCategoryWithData(categories.value);
    syncDraggableSiteGroups();
  }
}

const isMobileSidebarOpen = ref(false);
const toggleMobileSidebar = () => {
  isMobileSidebarOpen.value = !isMobileSidebarOpen.value;
}

onMounted(async () => {
  const timeout = setTimeout(() => {
    if (!isDataLoaded.value) connectionIssue.value = true;
  }, 5000);

  try {
    currentUser.value = await api.getMe();
    await fetchData();
    isDataLoaded.value = true;
    clearTimeout(timeout);
  } catch (e) {
    console.warn("Soft initialization warning:", e);
    isDataLoaded.value = true;
    clearTimeout(timeout);
  }

  const params = new URLSearchParams(window.location.search);
  if (params.get('sidebar_mode') === 'true') {
    isSidebarMode.value = true;
  }

  const extUrl = params.get('ext_url');
  const extTitle = params.get('ext_title');
  const quickSaveTargetUrl = params.get('target_url');
  const quickSaveTargetName = params.get('target_name');
  const quickSaveDescription = params.get('target_description');
  pendingQuickSave.value = params.get('quick_save') === 'true';

  if (quickSaveTargetUrl || extUrl) {
    currentPageData.value = {
      url: quickSaveTargetUrl || extUrl || '',
      name: quickSaveTargetName || extTitle || ''
    };
  } else {
    currentPageData.value = {
      url: window.location.href,
      name: document.title || '当前页面'
    };
  }

  if (quickSaveDescription && currentPageData.value) {
    currentPageData.value = {
      ...currentPageData.value,
      description: quickSaveDescription
    };
  }

  if (isSidebarMode.value && pendingQuickSave.value && currentPageData.value?.url) {
    await quickSaveCurrentPage();
  }
});

const handleLoginSuccess = async (user: any) => {
  currentUser.value = user;
  showLoginModal.value = false;
  await fetchData();
  if (pendingQuickSave.value && currentPageData.value?.url) {
    await quickSaveCurrentPage();
  }
}

const handleSelectCategory = (id: string | null) => {
  if (isAdminSeedRoute.value) {
    openBookmarkHome();
  }
  showOfficeToolbox.value = false;
  activeCategoryId.value = id;
  duplicateSaveMessage.value = '';
  showCompactCategoryPanel.value = false;
  if (window.innerWidth < 768) {
    isMobileSidebarOpen.value = false;
  }
}

const logout = async () => {
  try {
    await api.logout();
    currentUser.value = null;
    activeCategoryId.value = null;
    if (isAdminSeedRoute.value) {
      router.replace({ name: 'home' });
    }
    quickFilter.value = 'all';
    await fetchData();
  } catch (e) {
    currentUser.value = null;
  }
}

const openAddCategory = () => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  editingCategory.value = null;
  showCategoryForm.value = true;
}

const openEditCategory = (cat: Category) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  editingCategory.value = cat;
  showCategoryForm.value = true;
}

const saveCategory = async (data: any) => {
  if (!currentUser.value) return;
  editingCategory.value?.id
      ? await api.updateCategory(editingCategory.value.id, data)
      : await api.createCategory(data);
  showCategoryForm.value = false;
  await fetchData();
}

const deleteCategory = async (id: string) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  if (!confirm('确定删除分类及其中书签？')) return;
  await api.deleteCategory(id);
  await fetchData();
}

const handleUpdateCategoriesOrder = async (ids: string[]) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  try {
    await api.updateCategoriesOrder(ids);
    await fetchData();
  } catch (error: any) {
    console.error("Update categories order failed:", error);
  }
};

const openAddSite = () => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  editingSite.value = null;
  initialSiteData.value = null;
  preferredCategoryId.value = activeCategoryId.value || recentCategoryId.value || null;
  showSiteForm.value = true;
}

const openAddSiteWithCategory = (catId: string) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  activeCategoryId.value = catId;
  editingSite.value = null;
  initialSiteData.value = null;
  preferredCategoryId.value = catId;
  showSiteForm.value = true;
}

const openEditSite = (site: Site) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  editingSite.value = site;
  preferredCategoryId.value = site.categoryId;
  showSiteForm.value = true;
}

const handleSiteFormClose = () => {
  showSiteForm.value = false;
  initialSiteData.value = null;
  editingSite.value = null;
  preferredCategoryId.value = null;
  pendingQuickSave.value = false;
}

const saveSite = async (data: any) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  try {
    if (editingSite.value?.id) {
      await api.updateSite(editingSite.value.id, data);
    } else {
      await api.createSite(data);
    }
    window.localStorage.setItem(RECENT_CATEGORY_KEY, data.categoryId);
    duplicateSaveMessage.value = '';
    showSiteForm.value = false;
    pendingQuickSave.value = false;
    await fetchData();
  } catch (error: any) {
    if (error?.status === 409) {
      alert('这个网址已经保存过了，不需要重复添加。');
      return;
    }
    alert(error?.message || "保存失败");
  }
}

const deleteSite = async (id: string) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  if (!confirm('确定删除该书签？')) return;
  await api.deleteSite(id);
  await fetchData();
}

const handleOpenSite = async (site: Site) => {
  if (currentUser.value) {
    try {
      const data = await api.trackSiteVisit(site.id);
      sites.value = sites.value.map(existing => existing.id === site.id ? {
        ...existing,
        visitCount: data?.visitCount ?? ((existing.visitCount || 0) + 1),
        lastVisitedAt: data?.lastVisitedAt ?? new Date().toISOString(),
      } : existing);
    } catch (error) {
      console.error('Track site visit failed:', error);
    }
  }
  window.open(site.url, '_blank');
}

const toggleFavorite = async (site: Site) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  const nextFavorite = !site.isFavorite;
  await api.updateSiteState(site.id, { isFavorite: nextFavorite });
  sites.value = sites.value.map(existing => existing.id === site.id ? { ...existing, isFavorite: nextFavorite } : existing);
};

const cycleSiteStatus = async (site: Site) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }
  const nextStatus = cycleWorkflowStatusValue(site.workflowStatus);
  await api.updateSiteState(site.id, { workflowStatus: nextStatus });
  sites.value = sites.value.map(existing => existing.id === site.id ? { ...existing, workflowStatus: nextStatus } : existing);
};

watch([categories, sites], () => {
  syncDraggableSiteGroups();
}, { deep: true });

watch([isDataLoaded, currentUser, () => route.name], () => {
  if (!isDataLoaded.value) return;
  if (route.name === 'admin-seed' && !currentUser.value?.isSuperAdmin) {
    router.replace({ name: 'home' });
  }
});

watch([activeCategoryId, searchQuery, quickFilter], () => {
  visibleSiteLimit.value = SITE_PAGE_SIZE;
});

const filteredGroupedSites = computed(() => {
  const search = (searchQuery.value || '').trim().toLowerCase();
  const domainMatch = search.match(/domain:([^\s]+)/i);
  const tagMatch = search.match(/tag:([^\s]+)/i);
  const normalizedSearch = search
      .replace(/domain:[^\s]+/ig, '')
      .replace(/tag:[^\s]+/ig, '')
      .trim();

  let filtered = sites.value.filter(s => {
    if (domainMatch && !getHostname(s.url).includes(domainMatch[1].toLowerCase())) {
      return false;
    }
    if (tagMatch && !(s.tagsText || '').toLowerCase().includes(tagMatch[1].toLowerCase())) {
      return false;
    }
    if (!normalizedSearch) return true;
    const name = (s.name || '').toString().toLowerCase();
    const desc = (s.description || '').toString().toLowerCase();
    const url = (s.url || '').toString().toLowerCase();
    const tagsText = (s.tagsText || '').toString().toLowerCase();
    return name.includes(normalizedSearch) || desc.includes(normalizedSearch) || url.includes(normalizedSearch) || tagsText.includes(normalizedSearch);
  });

  if (quickFilter.value === 'favorites') {
    filtered = filtered.filter(site => !!site.isFavorite);
  } else if (quickFilter.value === 'read_later') {
    filtered = filtered.filter(site => site.workflowStatus === 'read_later');
  } else if (quickFilter.value === 'unorganized') {
    filtered = filtered.filter(site => site.workflowStatus === 'unorganized');
  } else if (quickFilter.value === 'recent_visited') {
    filtered = filtered.filter(site => !!site.lastVisitedAt);
  }

  if (activeCategoryId.value) {
    filtered = filtered.filter(s => String(s.categoryId) === String(activeCategoryId.value));
  }

  const groups: { category: Category, sites: Site[] }[] = [];
  const categorizedSiteIds = new Set<string>();

  categories.value.forEach(cat => {
    const sitesInCat = filtered.filter(s => String(s.categoryId) === String(cat.id));
    if (sitesInCat.length > 0) {
      groups.push({ category: cat, sites: sitesInCat });
      sitesInCat.forEach(s => categorizedSiteIds.add(s.id));
    }
  });

  const orphanSites = filtered.filter(s => !categorizedSiteIds.has(s.id));
  if (orphanSites.length > 0) {
    groups.push({
      category: { id: 'orphan', name: '未分类', iconName: '📁' } as Category,
      sites: orphanSites
    });
  }
  return groups.map(group => ({
    category: group.category,
    sites: [...group.sites].sort((a, b) => {
      if (quickFilter.value === 'recent_added') {
        return new Date(b.createdAt || 0).getTime() - new Date(a.createdAt || 0).getTime();
      }
      if (quickFilter.value === 'recent_visited') {
        const visitTimeDiff = new Date(b.lastVisitedAt || 0).getTime() - new Date(a.lastVisitedAt || 0).getTime();
        if (visitTimeDiff !== 0) return visitTimeDiff;
        return (b.visitCount || 0) - (a.visitCount || 0);
      }
      if (quickFilter.value === 'favorites') {
        return (b.visitCount || 0) - (a.visitCount || 0);
      }
      return (a.sortOrder ?? 0) - (b.sortOrder ?? 0);
    })
  }));
});

const visibleDraggableGroups = computed(() => {
  return draggableSiteGroups.value
      .filter(group => {
        if (activeCategoryId.value === null) {
          return group.sites.length > 0 || group.category.id === 'orphan';
        }
        return String(group.category.id) === String(activeCategoryId.value);
      });
});

const flatFilteredSites = computed(() => {
  return filteredGroupedSites.value.flatMap(group => group.sites);
});

const visibleFlatFilteredSites = computed(() => {
  return flatFilteredSites.value.slice(0, visibleSiteLimit.value);
});

const visibleFilteredGroupedSites = computed(() => {
  let remaining = visibleSiteLimit.value;
  return filteredGroupedSites.value
      .map(group => {
        if (remaining <= 0) {
          return { category: group.category, sites: [] };
        }
        const sites = group.sites.slice(0, remaining);
        remaining -= sites.length;
        return { category: group.category, sites };
      })
      .filter(group => group.sites.length > 0);
});

const hasMoreFilteredSites = computed(() => {
  return flatFilteredSites.value.length > visibleSiteLimit.value;
});

const loadMoreSites = () => {
  visibleSiteLimit.value += SITE_PAGE_SIZE;
};

const canDragSites = computed(() => {
  return !!currentUser.value
      && !searchQuery.value.trim()
      && quickFilter.value === 'all'
      && flatFilteredSites.value.length <= visibleSiteLimit.value;
});

const persistSiteGroups = async (groups: SiteGroup[]) => {
  const payload = groups
      .filter(group => group.category.id !== 'orphan')
      .flatMap(group => group.sites.map((site, index) => ({
        id: site.id,
        categoryId: String(group.category.id),
        sortOrder: index,
      })));

  await api.updateSitesOrder(payload);
};

const normalizeDraggedSites = () => {
  draggableSiteGroups.value = draggableSiteGroups.value.map(group => ({
    category: group.category,
    sites: group.sites.map((site, index) => ({
      ...site,
      categoryId: String(group.category.id),
      sortOrder: index,
    }))
  }));
};

const handleSitesDragStart = (event: { item: HTMLElement }) => {
  draggingSiteId.value = event.item.dataset.siteId || null;
  if (event.item.dataset.siteId) {
    event.item.classList.add('site-drag-source');
  }
};

const handleSitesDragEnd = async (event?: { item?: HTMLElement }) => {
  if (event?.item) {
    event.item.classList.remove('site-drag-source');
  }

  if (isDroppingToSidebarCategory.value) {
    draggingSiteId.value = null;
    isDroppingToSidebarCategory.value = false;
    return;
  }

  if (!currentUser.value) {
    openLogin();
    await fetchData();
    draggingSiteId.value = null;
    return;
  }

  normalizeDraggedSites();

  try {
    await persistSiteGroups(draggableSiteGroups.value);
    sites.value = draggableSiteGroups.value.flatMap(group => group.sites);
    syncDraggableSiteGroups();
  } catch (error) {
    console.error("Update sites order failed:", error);
    await fetchData();
  } finally {
    draggingSiteId.value = null;
  }
};

const handleMoveSiteToCategory = async ({ siteId, categoryId }: { siteId: string; categoryId: string }) => {
  if (!currentUser.value) {
    openLogin();
    return;
  }

  const sourceSite = sites.value.find(site => site.id === siteId);
  if (!sourceSite || String(sourceSite.categoryId) === String(categoryId)) {
    isDroppingToSidebarCategory.value = true;
    draggingSiteId.value = null;
    return;
  }

  isDroppingToSidebarCategory.value = true;

  const nextGroups = draggableSiteGroups.value
      .map(group => ({
        category: group.category,
        sites: group.sites.filter(site => site.id !== siteId),
      }))
      .filter(group => group.category.id !== 'orphan');

  const targetIndex = nextGroups.findIndex(group => String(group.category.id) === String(categoryId));
  if (targetIndex === -1) {
    draggingSiteId.value = null;
    isDroppingToSidebarCategory.value = false;
    return;
  }

  nextGroups[targetIndex] = {
    category: nextGroups[targetIndex].category,
    sites: [
      ...nextGroups[targetIndex].sites,
      { ...sourceSite, categoryId: String(categoryId) },
    ],
  };

  const normalizedGroups = nextGroups.map(group => ({
    category: group.category,
    sites: group.sites.map((site, index) => ({
      ...site,
      categoryId: String(group.category.id),
      sortOrder: index,
    })),
  }));

  try {
    await persistSiteGroups(normalizedGroups);
    sites.value = normalizedGroups.flatMap(group => group.sites);
    syncDraggableSiteGroups();
  } catch (error) {
    console.error("Move site to category failed:", error);
    await fetchData();
  } finally {
    draggingSiteId.value = null;
    isDroppingToSidebarCategory.value = false;
  }
};
</script>

<template>
  <div class="flex h-screen bg-[#F8FAFC] overflow-hidden font-body relative">

    <!-- 全局加载状态 -->
    <div v-if="!isDataLoaded" class="fixed inset-0 z-50 bg-[#F8FAFC] flex flex-col items-center justify-center">
      <Loader2 class="h-10 w-10 animate-spin text-blue-600 mb-6"/>
      <p class="text-xs font-bold text-slate-500 uppercase tracking-widest animate-pulse">Initializing System...</p>

      <div v-if="connectionIssue" class="mt-8 p-6 max-w-xs text-center border border-red-100 bg-red-50 rounded-2xl shadow-sm">
        <p class="text-sm text-red-600 font-bold mb-2">连接似乎遇到了问题</p>
        <p class="text-xs text-red-400/80 mb-6 leading-relaxed">这可能是由于网络延迟或服务暂时不可用。</p>
        <button @click="retryFetch" class="w-full py-3 bg-red-500 text-white rounded-xl text-sm font-bold shadow-sm shadow-red-200 active:scale-95 transition-all">重试连接</button>
      </div>
    </div>

    <!-- 移动端侧边栏遮罩 -->
    <div v-if="isMobileSidebarOpen" class="fixed inset-0 bg-slate-900/20 backdrop-blur-sm z-40 md:hidden" @click="isMobileSidebarOpen = false"></div>

    <!-- 侧边栏 (常规模式显示) -->
	    <aside
	        v-if="!isCompactWorkspace"
	        class="fixed inset-y-0 left-0 border-r border-slate-200 bg-white z-50 transform transition-[width,transform] duration-300 ease-out md:relative md:translate-x-0 shadow-[4px_0_24px_rgba(0,0,0,0.02)]"
	        :class="[
	          isMobileSidebarOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
	          isSidebarCollapsed ? 'w-64 md:w-20' : 'w-64'
	        ]"
	    >
	      <button
	          type="button"
	          class="absolute right-0 top-8 z-[60] hidden h-7 w-7 translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full border border-slate-200 bg-white text-slate-400 shadow-[0_8px_20px_rgba(15,23,42,0.10)] transition-all duration-200 hover:border-blue-200 hover:bg-blue-50 hover:text-blue-600 active:scale-95 md:flex"
	          :title="isSidebarCollapsed ? '展开分类栏' : '收起分类栏'"
	          :aria-label="isSidebarCollapsed ? '展开分类栏' : '收起分类栏'"
	          @click.stop="toggleSidebarCollapse"
	      >
	        <ChevronRight v-if="isSidebarCollapsed" class="h-4 w-4" />
	        <ChevronLeft v-else class="h-4 w-4" />
	      </button>
	      <div class="flex h-16 items-center border-b border-slate-200 transition-all duration-300 ease-out"
	           :class="isSidebarCollapsed ? 'px-3 justify-center overflow-hidden' : 'px-6 justify-start overflow-hidden'">
	        <div class="flex min-w-0 items-center gap-3 group cursor-pointer hover:opacity-80 transition-opacity"
	             :class="isSidebarCollapsed ? 'md:hidden' : ''"
	             @click="handleSelectCategory(null)">
          <div class="h-8 w-8 shrink-0 bg-gradient-to-br from-blue-500 to-cyan-500 rounded-xl flex items-center justify-center text-white shadow-md shadow-blue-100">
            <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z"/>
            </svg>
          </div>
          <span class="text-lg font-black text-slate-800 tracking-tight whitespace-nowrap overflow-hidden transition-all duration-300 ease-out"
                :class="isSidebarCollapsed ? 'md:w-0 md:opacity-0 md:translate-x-2' : 'w-auto opacity-100 translate-x-0'">
            LinkDock
          </span>
        </div>
	      </div>
      <Sidebar :categories="categories" :sites="sites" :activeCategoryId="activeCategoryId" :canEdit="true" :draggedSiteId="draggingSiteId"
               :collapsed="isSidebarCollapsed"
               @select-category="handleSelectCategory" @add-site="openAddSite" @add-category="openAddCategory"
               @edit-category="openEditCategory" @delete-category="deleteCategory"
               @update-categories-order="handleUpdateCategoriesOrder" @move-site-to-category="handleMoveSiteToCategory"/>
    </aside>

    <div class="flex-1 flex flex-col h-full overflow-hidden relative">
      <!-- 顶部 Header (常规模式显示) -->
	      <Header v-if="!isCompactWorkspace" v-model:searchQuery="searchQuery" :user="currentUser"
	              @login="openLogin" @logout="logout" @toggle-sidebar="toggleMobileSidebar" @add-site="openAddSite"
	              @open-toolbox="openOfficeToolbox" @change-password="openChangePassword" @manage-seed="openAdminSeedManager"/>

      <!-- 扩展侧边栏专用极简 Header -->
      <div v-if="isCompactWorkspace" class="relative border-b border-slate-200/80 flex flex-col shrink-0 bg-[linear-gradient(180deg,#ffffff_0%,#f8fbff_100%)]">
        <div class="px-4 pt-4 pb-3">
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <div class="flex items-center gap-2">
                <div class="flex h-8 w-8 items-center justify-center rounded-xl bg-slate-900 text-white shadow-sm shadow-slate-200">
                  <svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
                    <path d="M13 10V3L4 14h7v7l9-11h-7z"/>
                  </svg>
                </div>
                <div>
                  <p class="text-[11px] font-black uppercase tracking-[0.24em] text-slate-800">LinkDock</p>
                  <p class="mt-0.5 text-[11px] text-slate-400">
                    {{ currentUser ? `你好，${currentUser.username}` : '浏览你的扩展工作台' }}
                  </p>
                </div>
              </div>
            </div>

            <div class="flex shrink-0 items-center gap-2">
              <button
                  v-if="currentUser"
                  @click="logout"
                  class="inline-flex h-8 items-center rounded-xl border border-slate-200 bg-white px-3 text-[11px] font-bold text-slate-500 transition-all hover:border-slate-300 hover:text-slate-900"
              >
                退出
              </button>
              <button
                  v-else
                  @click="openLogin()"
                  class="inline-flex h-8 items-center rounded-xl border border-slate-200 bg-white px-3 text-[11px] font-bold text-slate-500 transition-all hover:border-slate-300 hover:text-slate-900"
              >
                登录
              </button>
            </div>
          </div>
        </div>

        <div v-if="isSidebarMode && duplicateSaveMessage && !showSiteForm" class="px-4 pb-2">
          <div class="flex items-start gap-3 rounded-2xl border border-amber-200 bg-amber-50 px-3 py-2.5 shadow-sm">
            <div class="mt-0.5 flex h-6 w-6 shrink-0 items-center justify-center rounded-xl bg-white text-amber-500 shadow-sm">
              <AlertTriangle class="h-3.5 w-3.5" />
            </div>
            <div class="min-w-0 flex-1">
              <p class="text-xs font-black text-amber-700">已保存该网址</p>
              <p class="mt-0.5 text-[11px] leading-relaxed text-amber-700/80">{{ duplicateSaveMessage }}</p>
            </div>
            <button
                type="button"
                @click="duplicateSaveMessage = ''"
                class="shrink-0 rounded-lg px-1.5 py-1 text-[10px] font-bold text-amber-700/70 transition-colors hover:bg-white hover:text-amber-700"
            >
              知道了
            </button>
          </div>
        </div>

        <div v-if="isSidebarMode && currentPageData?.url && !currentPageExistingSite && !showSiteForm" class="px-4 pb-2">
          <div class="flex items-center gap-3 rounded-2xl border border-slate-200 bg-white/90 px-3 py-2 shadow-sm">
            <div class="min-w-0 flex-1">
              <div class="flex items-center gap-2">
                <p class="truncate text-xs font-bold text-slate-800">{{ currentPageData?.name || currentPageData?.url }}</p>
                <span
                    v-if="currentPageSameDomainSites.length > 0"
                    class="shrink-0 rounded-full bg-blue-50 px-2 py-0.5 text-[10px] font-bold text-blue-600"
                >
                  同域名
                </span>
              </div>
              <p class="mt-0.5 truncate text-[10px] text-slate-400">
                <span v-if="recommendedCategoryForCurrentPage">推荐放入 {{ recommendedCategoryForCurrentPage.name }}</span>
                <span v-else>{{ currentPageData?.url }}</span>
              </p>
            </div>
            <button
                @click="openSaveCurrentPage"
                class="shrink-0 rounded-xl bg-blue-600 px-3 py-2 text-[11px] font-bold text-white shadow-sm shadow-blue-200/70 transition-all hover:bg-blue-700"
            >
              保存
            </button>
          </div>
        </div>

        <div v-if="!showSiteForm" class="px-4 pb-3">
          <div class="relative group">
            <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-slate-400" />
            <input
                v-model="searchQuery"
                placeholder="搜索书签、域名、标签..."
                class="w-full h-10 pl-8 pr-4 bg-white/90 border border-slate-200 rounded-2xl text-xs shadow-sm focus:outline-none focus:border-blue-500 focus:ring-4 focus:ring-blue-500/10 transition-all"
            />
          </div>
        </div>
        <div v-if="!showSiteForm" class="px-4 pb-3">
          <div class="flex items-center gap-2">
            <button
                v-if="categories.length > 0"
                @click="showCompactCategoryPanel = !showCompactCategoryPanel; showCompactFilterPanel = false"
                class="flex min-w-0 flex-1 items-center justify-between rounded-2xl border border-slate-200 bg-white px-3 py-2 text-left shadow-sm transition-all hover:border-slate-300"
            >
              <div class="min-w-0">
                <p class="text-[9px] font-bold uppercase tracking-[0.18em] text-slate-400">分类</p>
                <p class="truncate text-[11px] font-bold text-slate-700">
                  {{ activeCategoryId ? categories.find(cat => String(cat.id) === activeCategoryId)?.name : '全部书签' }}
                </p>
              </div>
              <ChevronDown class="ml-1 h-3.5 w-3.5 shrink-0 text-slate-400 transition-transform" :class="showCompactCategoryPanel ? 'rotate-180' : ''" />
            </button>

            <button
                @click="showCompactFilterPanel = !showCompactFilterPanel; showCompactCategoryPanel = false"
                class="flex min-w-0 flex-1 items-center justify-between rounded-2xl border border-slate-200 bg-white px-3 py-2 text-left shadow-sm transition-all hover:border-slate-300"
            >
              <div class="min-w-0">
                <p class="text-[9px] font-bold uppercase tracking-[0.18em] text-slate-400">筛选</p>
                <p class="truncate text-[11px] font-bold text-slate-700">{{ activeQuickFilterLabel }}</p>
              </div>
              <ChevronDown class="ml-1 h-3.5 w-3.5 shrink-0 text-slate-400 transition-transform" :class="showCompactFilterPanel ? 'rotate-180' : ''" />
            </button>
          </div>
        </div>

        <Transition name="compact-popover">
          <div v-if="!showSiteForm && showCompactCategoryPanel && categories.length > 0" class="absolute left-4 right-4 top-full z-30 -mt-2 rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_18px_42px_-20px_rgba(15,23,42,0.32)]">
            <div class="mb-2 flex items-center justify-between">
              <p class="text-[10px] font-bold uppercase tracking-[0.22em] text-slate-400">分类</p>
              <span class="text-[10px] font-semibold text-slate-300">{{ categories.length + 1 }} 项</span>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                  @click="activeCategoryId = null; showCompactCategoryPanel = false"
                  class="whitespace-nowrap px-2.5 py-1.5 rounded-full text-[10px] font-bold transition-all"
                  :class="activeCategoryId === null ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'"
              >
                全部
              </button>
              <button
                  v-for="cat in categories"
                  :key="cat.id"
                  @click="activeCategoryId = String(cat.id); showCompactCategoryPanel = false"
                  class="whitespace-nowrap px-2.5 py-1.5 rounded-full text-[10px] font-bold transition-all"
                  :class="activeCategoryId === String(cat.id) ? 'bg-slate-900 text-white' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'"
              >
                {{ cat.iconName }} {{ cat.name }}
              </button>
            </div>
          </div>
        </Transition>

        <Transition name="compact-popover">
          <div v-if="!showSiteForm && showCompactFilterPanel" class="absolute left-4 right-4 top-full z-30 -mt-2 rounded-2xl border border-slate-200 bg-white p-3 shadow-[0_18px_42px_-20px_rgba(15,23,42,0.32)]">
            <div class="mb-2 flex items-center justify-between">
              <p class="text-[10px] font-bold uppercase tracking-[0.22em] text-slate-400">筛选</p>
              <span class="text-[10px] font-semibold text-slate-300">快速视图</span>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                  v-for="filter in quickFilters"
                  :key="filter.id"
                  @click="quickFilter = filter.id; showCompactFilterPanel = false"
                  class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1.5 text-[10px] font-bold transition-all"
                  :class="quickFilter === filter.id ? 'bg-blue-600 text-white shadow-sm shadow-blue-200/70' : 'bg-slate-100 text-slate-500 hover:bg-slate-200'"
              >
                <component :is="filter.icon" class="h-3 w-3" />
                <span>{{ filter.label }}</span>
              </button>
            </div>
          </div>
        </Transition>
      </div>

      <!-- 主体内容区 -->
	      <main class="flex-1 overflow-y-auto relative custom-scrollbar" :class="canShowAdminSeedPage ? 'p-0' : isCompactWorkspace ? 'p-4' : 'p-6 md:p-10'">
	        <AdminSeedManager
	            v-if="canShowAdminSeedPage"
	            @close="openBookmarkHome"
	            @saved="fetchData"
	        />
	        <OfficeToolbox
	            v-else-if="showOfficeToolbox"
	            :currentPageData="currentPageData"
	        />
	
	        <div v-else class="max-w-6xl mx-auto w-full">

          <div v-if="!isCompactWorkspace" class="mb-8 flex flex-col gap-3 pb-6 border-b border-slate-200/50 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex flex-wrap items-center gap-2">
              <button
                  v-for="filter in quickFilters"
                  :key="filter.id"
                  @click="quickFilter = filter.id"
                  class="inline-flex items-center gap-1.5 rounded-full px-3 py-1.5 text-xs font-bold transition-all"
                  :class="quickFilter === filter.id ? 'bg-blue-600 text-white shadow-md shadow-blue-200/60' : 'bg-white text-slate-500 border border-slate-200 hover:border-slate-300'"
              >
                <component :is="filter.icon" class="h-3.5 w-3.5" />
                <span>{{ filter.label }}</span>
              </button>
            </div>
            <button
                v-if="canEdit"
                @click="activeCategoryId ? openAddSiteWithCategory(activeCategoryId) : openAddSite()"
                class="inline-flex h-9 w-fit shrink-0 items-center gap-2 rounded-xl bg-blue-600 px-4 text-xs font-bold text-white shadow-md shadow-blue-200/60 transition-all hover:-translate-y-0.5 hover:bg-blue-700 active:scale-[0.97]"
                title="添加书签"
            >
              <Plus class="h-4 w-4" />
              <span>添加书签</span>
            </button>
          </div>

          <!-- 正常内容展示 -->
          <div class="w-full">

            <div v-if="flatFilteredSites.length > 0" class="space-y-12">

              <div v-if="activeCategoryId === null" class="space-y-4">
                <div class="flex items-center group/header cursor-default mb-2.5">
                  <span class="text-xl shrink-0 mr-3">📚</span>
                  <h3 class="text-base font-bold text-slate-700 tracking-tight mr-4">所有书签</h3>

                  <div class="h-[1px] flex-1 bg-gradient-to-r from-slate-200 to-transparent"></div>

	                  <div class="flex items-center gap-2.5 ml-4">
	                    <span class="text-[10px] text-slate-400 font-bold uppercase tracking-widest bg-white px-2.5 py-0.5 rounded-full border border-slate-200">
	                      {{ visibleFlatFilteredSites.length }} / {{ flatFilteredSites.length }} ITEMS
	                    </span>
	                  </div>
	                </div>

                <div v-if="canDragSites" class="space-y-8">
                  <div v-for="group in visibleDraggableGroups" :key="group.category.id" class="space-y-4">
                    <div class="flex items-center gap-3">
                      <span class="text-lg shrink-0">{{ group.category.iconName }}</span>
                      <h4 class="text-sm font-bold text-slate-600">{{ group.category.name }}</h4>
                      <div class="h-px flex-1 bg-gradient-to-r from-slate-200 to-transparent"></div>
                      <span class="text-[10px] font-bold uppercase tracking-widest text-slate-400">{{ group.sites.length }} items</span>
                    </div>

                    <draggable
                        v-model="group.sites"
                        item-key="id"
                        :group="{ name: 'sites' }"
                        :class="isSidebarMode ? 'grid gap-3 min-h-[1.5rem] [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]' : 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-5 min-h-[1.5rem]'"
                        ghost-class="site-drag-ghost"
                        drag-class="site-drag-active"
                        @start="handleSitesDragStart"
                        @end="handleSitesDragEnd"
                    >
                      <template #item="{ element }">
                        <SiteCard :site="element" :canEdit="true" :compact="isSidebarMode" @open="handleOpenSite" @toggle-favorite="toggleFavorite" @cycle-status="cycleSiteStatus" @edit="openEditSite" @delete="deleteSite"/>
                      </template>
                    </draggable>
                  </div>
                </div>

	                <div v-else :class="isSidebarMode ? 'grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]' : 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-5'">
	                  <SiteCard v-for="site in visibleFlatFilteredSites" :key="site.id" :site="site" :canEdit="true" :compact="isSidebarMode"
	                            @open="handleOpenSite" @toggle-favorite="toggleFavorite" @cycle-status="cycleSiteStatus"
	                            @edit="openEditSite" @delete="deleteSite"/>
	                </div>
              </div>

	              <div v-else class="space-y-4">
	                <div v-for="group in (canDragSites ? visibleDraggableGroups : visibleFilteredGroupedSites)" :key="group.category.id">
	                  <draggable
	                      v-if="canDragSites"
	                      v-model="group.sites"
	                      item-key="id"
	                      :class="isSidebarMode ? 'grid gap-3 min-h-[1.5rem] [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]' : 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-5 min-h-[1.5rem]'"
	                      ghost-class="site-drag-ghost"
                      drag-class="site-drag-active"
                      @start="handleSitesDragStart"
                      @end="handleSitesDragEnd"
                  >
                    <template #item="{ element }">
                      <SiteCard :site="element" :canEdit="true" :compact="isSidebarMode" @open="handleOpenSite" @toggle-favorite="toggleFavorite" @cycle-status="cycleSiteStatus" @edit="openEditSite" @delete="deleteSite"/>
                    </template>
                  </draggable>

	                  <div v-else :class="isSidebarMode ? 'grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]' : 'grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-5'">
	                    <SiteCard v-for="site in group.sites" :key="site.id" :site="site" :canEdit="true" :compact="isSidebarMode"
	                              @open="handleOpenSite" @toggle-favorite="toggleFavorite" @cycle-status="cycleSiteStatus"
	                              @edit="openEditSite" @delete="deleteSite"/>
	                  </div>
	                </div>
	              </div>

	              <div v-if="hasMoreFilteredSites" class="flex justify-center pt-2">
	                <button
	                    @click="loadMoreSites"
	                    class="inline-flex h-10 items-center justify-center rounded-xl border border-slate-200 bg-white px-5 text-xs font-bold text-slate-600 shadow-sm transition-all hover:-translate-y-0.5 hover:border-blue-100 hover:text-blue-600 hover:shadow-md active:scale-[0.97]"
	                >
	                  加载更多 {{ Math.min(SITE_PAGE_SIZE, flatFilteredSites.length - visibleSiteLimit) }} 个
	                </button>
	              </div>
	            </div>

            <!-- 空状态 -->
            <div v-else-if="!isCompactWorkspace" class="flex flex-col items-center justify-center py-32 bg-white rounded-3xl border border-slate-200 shadow-sm">
              <div class="w-20 h-20 bg-slate-50 rounded-2xl flex items-center justify-center mb-6">
                <SearchX class="h-10 w-10 text-slate-300"/>
              </div>
              <p class="text-xs font-bold text-slate-400 uppercase tracking-widest mb-6">暂时没有发现任何内容</p>

	              <button
	                  v-if="canEdit && currentUser"
	                  @click="activeCategoryId ? openAddSiteWithCategory(activeCategoryId) : openAddSite()"
	                  class="flex items-center gap-2 px-6 h-10 rounded-lg bg-blue-600 text-white font-bold text-sm shadow-md shadow-blue-200/50 hover:bg-blue-700 hover:-translate-y-px transition-all active:scale-[0.97]"
	              >
                <Plus class="w-4 h-4" />
                <span>添加第一个书签</span>
              </button>
            </div>

            <div v-else-if="isCompactWorkspace && !showSiteForm" class="rounded-3xl border border-slate-200 bg-white px-5 py-10 text-center shadow-sm">
              <div class="mx-auto mb-4 flex h-14 w-14 items-center justify-center rounded-2xl bg-slate-50">
                <SearchX class="h-6 w-6 text-slate-300"/>
              </div>
              <h3 class="text-sm font-black text-slate-800">
                {{ activeCategoryId ? '这个分类还是空的' : '这里暂时还没有书签' }}
              </h3>
              <p class="mt-2 text-xs leading-relaxed text-slate-400">
                {{ currentUser ? '现在就添加一个书签，让这个分类开始工作。' : '你可以先浏览默认书签；想添加自己的内容时，再登录即可。' }}
              </p>
              <button
                  @click="openAddSite"
                  class="mt-5 inline-flex h-9 items-center justify-center rounded-xl bg-blue-600 px-4 text-xs font-bold text-white shadow-md shadow-blue-200/60 transition-all hover:-translate-y-0.5 hover:bg-blue-700"
              >
                {{ currentUser ? '添加书签' : '登录后添加书签' }}
              </button>
            </div>

          </div>

        </div>
      </main>

      <!-- 底部 Footer -->
      <footer v-if="!isCompactWorkspace" class="px-6 md:px-10 py-5 border-t border-slate-200 bg-white text-[11px] font-medium text-slate-500">
        <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div class="leading-relaxed">
<!--            <span class="font-bold text-slate-700">LinkDock</span>-->
<!--            <span class="mx-1.5 text-slate-300">/</span>-->
            <span>由个人开发的浏览器书签工作台，欢迎关注和 Star。</span>
          </div>
          <a
              :href="OPEN_SOURCE_URL"
              target="_blank"
              rel="noreferrer"
              class="inline-flex w-fit items-center gap-2 rounded-full border border-slate-200 bg-slate-50 px-3 py-2 text-xs font-bold text-slate-700 transition-all hover:-translate-y-0.5 hover:border-slate-300 hover:bg-white hover:text-blue-600 hover:shadow-sm"
          >
            <Github class="h-4 w-4" />
            <span>GitHub 开源地址</span>
          </a>
        </div>
      </footer>
    </div>

    <!-- 弹窗组件 -->
    <CategoryForm :isOpen="showCategoryForm" :category="editingCategory" @close="showCategoryForm = false"
                  @save="saveCategory"/>
    <SiteForm :isOpen="showSiteForm" :site="editingSite" :categories="categories" :sites="sites"
              :initialCategoryId="preferredCategoryId || activeCategoryId" :initialData="initialSiteData"
              @close="handleSiteFormClose" @save="saveSite" @edit-existing="openEditSite"/>
    <LoginModal :isOpen="showLoginModal" :initialIsRegister="initialIsRegister" @close="showLoginModal = false" @login="handleLoginSuccess"/>
    <ChangePasswordModal :isOpen="showChangePasswordModal" @close="showChangePasswordModal = false" @success="showChangePasswordModal = false" />

    <!-- 全局快捷搜索面板 -->
    <CommandPalette :isOpen="showCommandPalette" :sites="sites" :categories="categories"
                    @close="showCommandPalette = false" @selectSite="handleCommandSelectSite"
                    @selectCategory="handleCommandSelectCategory"/>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(203, 213, 225, 0.6);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(148, 163, 184, 0.8);
}

.site-drag-ghost {
  opacity: 0.45;
}

.site-drag-active {
  transform: rotate(2deg);
}

:deep(.site-drag-source) {
  opacity: 0.2;
}

.compact-popover-enter-active,
.compact-popover-leave-active {
  transition: opacity 140ms ease, transform 140ms ease;
}

.compact-popover-enter-from,
.compact-popover-leave-to {
  opacity: 0;
  transform: translateY(-6px) scale(0.98);
}
</style>
