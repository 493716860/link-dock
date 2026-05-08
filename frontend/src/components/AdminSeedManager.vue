<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue';
import { api } from '@/api';
import type { SeedCategory, SeedSite } from '@/types';
import { AlertTriangle, ArrowDown, ArrowLeft, ArrowUp, Database, Loader2, Plus, RefreshCw, Save, Trash2, Wand2 } from 'lucide-vue-next';

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'saved'): void;
}>();

const categories = ref<SeedCategory[]>([]);
const sites = ref<SeedSite[]>([]);
const isLoading = ref(false);
const isSaving = ref(false);
const errorMessage = ref('');
const successMessage = ref('');
const highlightedCategoryId = ref('');
const highlightedSiteId = ref('');
const openIconPickerCategoryId = ref('');
let highlightTimer: ReturnType<typeof window.setTimeout> | null = null;

const iconOptions = [
  '📁', '🤖', '👨‍💻', '🛠️', '🎬', '🎨', '📚', '⭐',
  '🔥', '💡', '🚀', '🧭', '🌍', '📰', '🛒', '💬',
  '🎵', '🎮', '☕', '📦', '📎', '📊', '⚙️', '🔒',
  '☁️', '💰', '🎓', '📝', '📌', '🧰', '✨', '🎯',
];

const slugify = (value: string, fallback: string) => {
  const slug = value
      .trim()
      .toLowerCase()
      .replace(/https?:\/\//g, '')
      .replace(/[^a-z0-9]+/g, '_')
      .replace(/^_+|_+$/g, '');
  return slug || fallback;
};

const nextCategoryId = () => `cat_${Date.now().toString(36)}`;
const nextSiteId = (name = '') => `site_${slugify(name, Date.now().toString(36))}`;
const defaultCategory = (): SeedCategory => ({
  id: 'default',
  name: '默认分类',
  iconName: '📁',
});

const scrollToSeedCard = async (type: 'category' | 'site', id: string) => {
  await nextTick();
  if (highlightTimer) window.clearTimeout(highlightTimer);
  highlightedCategoryId.value = type === 'category' ? id : '';
  highlightedSiteId.value = type === 'site' ? id : '';

  const escapedId = window.CSS?.escape ? window.CSS.escape(id) : id.replace(/"/g, '\\"');
  const card = document.querySelector<HTMLElement>(`[data-seed-${type}-id="${escapedId}"]`);
  if (!card) return;

  card.scrollIntoView({ behavior: 'smooth', block: 'center' });
  const firstInput = card.querySelector<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>('input, textarea, select');
  window.setTimeout(() => {
    firstInput?.focus();
    if (firstInput instanceof HTMLInputElement || firstInput instanceof HTMLTextAreaElement) {
      firstInput.select();
    }
  }, 320);

  highlightTimer = window.setTimeout(() => {
    highlightedCategoryId.value = '';
    highlightedSiteId.value = '';
    highlightTimer = null;
  }, 2400);
};

const categoryIdSet = computed(() => new Set(categories.value.map(category => category.id)));

const sitesByCategory = computed(() => {
  const grouped = new Map<string, number>();
  sites.value.forEach(site => {
    grouped.set(site.categoryId, (grouped.get(site.categoryId) || 0) + 1);
  });
  return grouped;
});

const validationError = computed(() => {
  if (categories.value.length === 0) return '至少需要保留一个默认分类';

  const categoryIds = new Set<string>();
  for (const category of categories.value) {
    if (!category.id.trim()) return '分类 ID 不能为空';
    if (!category.name.trim()) return '分类名称不能为空';
    if (categoryIds.has(category.id.trim())) return `分类 ID 重复：${category.id}`;
    categoryIds.add(category.id.trim());
  }

  const siteIds = new Set<string>();
  for (const site of sites.value) {
    if (!site.id.trim()) return '书签 ID 不能为空';
    if (!site.name.trim()) return '书签名称不能为空';
    if (!site.url.trim()) return `书签「${site.name || site.id}」的网址不能为空`;
    if (!categoryIds.has(site.categoryId)) return `书签「${site.name || site.id}」还没有选择有效分类`;
    if (siteIds.has(site.id.trim())) return `书签 ID 重复：${site.id}`;
    siteIds.add(site.id.trim());
  }

  return '';
});

const loadSeedData = async () => {
  isLoading.value = true;
  errorMessage.value = '';
  successMessage.value = '';
  try {
    const data = await api.getSeedData();
    const loadedCategories = (data.categories || []).map((category: SeedCategory) => ({ ...category }));
    categories.value = loadedCategories.some(category => category.id === 'default')
        ? loadedCategories
        : [defaultCategory(), ...loadedCategories];
    sites.value = (data.sites || [])
        .map((site: SeedSite, index: number) => ({ ...site, sortOrder: site.sortOrder ?? index, originalIndex: index }))
        .sort((a: SeedSite & { originalIndex: number }, b: SeedSite & { originalIndex: number }) => (a.sortOrder ?? a.originalIndex) - (b.sortOrder ?? b.originalIndex))
        .map(({ originalIndex, ...site }: SeedSite & { originalIndex: number }) => site);
  } catch (error: any) {
    errorMessage.value = error?.message || '默认数据加载失败';
  } finally {
    isLoading.value = false;
  }
};

onMounted(loadSeedData);

const addCategory = () => {
  const category = {
    id: nextCategoryId(),
    name: '新分类',
    iconName: '📁',
  };
  categories.value.push(category);
  errorMessage.value = '';
  successMessage.value = '';
  scrollToSeedCard('category', category.id);
};

const moveCategory = (index: number, direction: -1 | 1) => {
  const nextIndex = index + direction;
  if (nextIndex < 0 || nextIndex >= categories.value.length) return;
  const nextCategories = [...categories.value];
  [nextCategories[index], nextCategories[nextIndex]] = [nextCategories[nextIndex], nextCategories[index]];
  categories.value = nextCategories;
};

const deleteCategory = (categoryId: string) => {
  if (categories.value.length <= 1) {
    errorMessage.value = '至少需要保留一个默认分类';
    return;
  }
  if (sites.value.some(site => site.categoryId === categoryId)) {
    errorMessage.value = '请先移动或删除该分类下的书签';
    return;
  }
  categories.value = categories.value.filter(category => category.id !== categoryId);
};

const selectCategoryIcon = (category: SeedCategory, icon: string) => {
  category.iconName = icon;
  openIconPickerCategoryId.value = '';
};

const addSite = (categoryId?: string) => {
  const fallbackCategoryId = categoryId || categories.value[0]?.id || '';
  const site = {
    id: nextSiteId(),
    categoryId: fallbackCategoryId,
    name: '新书签',
    url: 'https://example.com',
    description: '',
    icon: '',
    sortOrder: sites.value.length,
  };
  sites.value.push(site);
  errorMessage.value = '';
  successMessage.value = '';
  scrollToSeedCard('site', site.id);
};

const deleteSite = (siteId: string) => {
  sites.value = sites.value.filter(site => site.id !== siteId);
};

const moveSite = (index: number, direction: -1 | 1) => {
  const nextIndex = index + direction;
  if (nextIndex < 0 || nextIndex >= sites.value.length) return;
  const nextSites = [...sites.value];
  [nextSites[index], nextSites[nextIndex]] = [nextSites[nextIndex], nextSites[index]];
  sites.value = nextSites.map((site, sortOrder) => ({ ...site, sortOrder }));
};

const fillSiteMetadata = async (site: SeedSite) => {
  if (!site.url.trim()) return;
  errorMessage.value = '';
  try {
    const meta = await api.fetchUrlMetadata(site.url);
    if (meta?.success) {
      if (!site.name.trim() || site.name === '新书签') site.name = meta.title || site.name;
      if (!site.description.trim()) site.description = meta.description || site.description;
      site.icon = meta.icon || site.icon;
      site.id = site.id || nextSiteId(site.name);
    }
  } catch (error: any) {
    errorMessage.value = error?.message || '获取网站信息失败';
  }
};

const saveSeedData = async () => {
  errorMessage.value = validationError.value;
  successMessage.value = '';
  if (errorMessage.value) return;

  isSaving.value = true;
  try {
    await api.saveSeedData({
      categories: categories.value.map(category => ({
        id: category.id.trim(),
        name: category.name.trim(),
        iconName: category.iconName.trim() || '📁',
      })),
      sites: sites.value.map((site, index) => ({
        id: site.id.trim(),
        categoryId: site.categoryId,
        name: site.name.trim(),
        url: site.url.trim(),
        description: site.description.trim(),
        icon: (site.icon || '').trim(),
        sortOrder: index,
      })),
    });
    successMessage.value = '默认数据已保存。未登录公共书签会立即生效，新用户注册时也会使用这份数据。';
    emit('saved');
  } catch (error: any) {
    errorMessage.value = error?.message || '保存默认数据失败';
  } finally {
    isSaving.value = false;
  }
};
</script>

<template>
  <div class="mx-auto w-full max-w-7xl">
    <div class="mb-6 overflow-hidden rounded-[2rem] border border-slate-200 bg-[radial-gradient(circle_at_top_left,#dbeafe_0%,transparent_34%),linear-gradient(135deg,#ffffff_0%,#f8fafc_100%)] p-6 shadow-sm">
      <div class="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
        <div class="min-w-0">
          <button
              type="button"
              @click="emit('close')"
              class="mb-5 inline-flex h-9 items-center gap-2 rounded-full border border-slate-200 bg-white/80 px-3 text-xs font-bold text-slate-500 transition-all hover:-translate-y-0.5 hover:border-slate-300 hover:text-slate-900"
          >
            <ArrowLeft class="h-3.5 w-3.5" />
            返回书签主页
          </button>
          <div class="flex items-center gap-3">
            <div class="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-slate-900 text-white shadow-lg shadow-slate-200">
              <Database class="h-5 w-5" />
            </div>
            <div>
              <p class="text-[11px] font-black uppercase tracking-[0.28em] text-blue-600">Super Admin</p>
              <h1 class="mt-1 text-3xl font-black tracking-tight text-slate-900">默认书签管理</h1>
            </div>
          </div>
          <p class="mt-4 max-w-2xl text-sm leading-relaxed text-slate-500">
            这里维护的是未登录用户看到的公共默认书签，也是新用户注册后初始化的数据模板。保存后会直接写回后端默认数据文件。
          </p>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <button
              type="button"
              @click="loadSeedData"
              class="inline-flex h-10 items-center gap-2 rounded-2xl border border-slate-200 bg-white px-4 text-xs font-bold text-slate-600 shadow-sm transition-all hover:-translate-y-0.5 hover:border-slate-300"
          >
            <RefreshCw class="h-3.5 w-3.5" />
            重新加载
          </button>
          <button
              type="button"
              :disabled="isSaving || !!validationError"
              @click="saveSeedData"
              class="inline-flex h-10 items-center gap-2 rounded-2xl bg-blue-600 px-5 text-xs font-bold text-white shadow-lg shadow-blue-200/70 transition-all hover:-translate-y-0.5 hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50 disabled:hover:translate-y-0"
          >
            <Loader2 v-if="isSaving" class="h-3.5 w-3.5 animate-spin" />
            <Save v-else class="h-3.5 w-3.5" />
            保存默认数据
          </button>
        </div>
      </div>
    </div>

    <div v-if="isLoading" class="flex h-80 flex-col items-center justify-center rounded-[2rem] border border-slate-200 bg-white text-slate-400 shadow-sm">
      <Loader2 class="mb-3 h-8 w-8 animate-spin text-blue-600" />
      <p class="text-sm font-bold">正在加载默认数据...</p>
    </div>

    <div v-else class="space-y-5">
      <div v-if="errorMessage || validationError" class="flex items-start gap-3 rounded-3xl border border-red-100 bg-red-50 px-5 py-4 text-sm font-bold text-red-600">
        <AlertTriangle class="mt-0.5 h-4 w-4 shrink-0" />
        <span>{{ errorMessage || validationError }}</span>
      </div>
      <div v-if="successMessage" class="rounded-3xl border border-emerald-100 bg-emerald-50 px-5 py-4 text-sm font-bold text-emerald-700">
        {{ successMessage }}
      </div>

      <section class="rounded-[2rem] border border-slate-200 bg-white p-5 shadow-sm">
        <div class="mb-5 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 class="text-lg font-black text-slate-900">默认分类</h2>
            <p class="mt-1 text-xs text-slate-400">分类 ID 会被书签引用。修改已有 ID 时，请同步检查书签所属分类。</p>
          </div>
          <button
              type="button"
              @click="addCategory"
              class="inline-flex h-10 items-center justify-center gap-2 rounded-2xl bg-slate-900 px-4 text-xs font-bold text-white transition-all hover:-translate-y-0.5"
          >
            <Plus class="h-3.5 w-3.5" />
            新增分类
          </button>
        </div>

        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <div
              v-for="(category, index) in categories"
              :key="category.id"
              :data-seed-category-id="category.id"
              class="rounded-3xl border p-4 transition-all duration-300 hover:border-slate-200 hover:bg-white hover:shadow-sm"
              :class="highlightedCategoryId === category.id ? 'border-blue-300 bg-blue-50/70 shadow-lg shadow-blue-100/70 ring-4 ring-blue-500/10' : 'border-slate-100 bg-slate-50/80'"
          >
            <div class="mb-4 flex items-center justify-between gap-3">
              <div class="flex items-center gap-3">
                <div class="flex h-11 w-11 items-center justify-center rounded-2xl bg-white text-xl shadow-sm ring-1 ring-slate-100">
                  {{ category.iconName || '📁' }}
                </div>
                <div class="min-w-0">
                  <p class="truncate text-sm font-black text-slate-800">{{ category.name || '未命名分类' }}</p>
                  <p class="mt-0.5 text-xs font-semibold text-slate-400">{{ sitesByCategory.get(category.id) || 0 }} 个书签</p>
                </div>
              </div>
              <div class="flex shrink-0 items-center gap-1.5">
                <button
                    type="button"
                    :disabled="index === 0"
                    @click="moveCategory(index, -1)"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-400 transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600 disabled:cursor-not-allowed disabled:opacity-40"
                    title="上移分类"
                >
                  <ArrowUp class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    :disabled="index === categories.length - 1"
                    @click="moveCategory(index, 1)"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-400 transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600 disabled:cursor-not-allowed disabled:opacity-40"
                    title="下移分类"
                >
                  <ArrowDown class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    @click="addSite(category.id)"
                    class="inline-flex h-8 items-center rounded-xl bg-blue-50 px-2.5 text-xs font-bold text-blue-600 transition-all hover:bg-blue-600 hover:text-white"
                >
                  加书签
                </button>
              </div>
            </div>

	            <div class="space-y-2">
	              <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">图标</label>
	              <div class="rounded-2xl border border-slate-200 bg-white p-3">
	                <div class="flex items-center gap-2">
	                  <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-slate-50 text-xl ring-1 ring-slate-100">
	                    {{ category.iconName || '📁' }}
	                  </div>
	                  <input
	                      v-model="category.iconName"
	                      class="h-10 min-w-0 flex-1 rounded-xl border border-slate-200 bg-slate-50 px-3 text-sm font-bold outline-none focus:border-blue-300 focus:bg-white focus:ring-4 focus:ring-blue-500/10"
	                      placeholder="也可以手动输入 emoji"
	                  />
	                  <button
	                      type="button"
	                      @click="openIconPickerCategoryId = openIconPickerCategoryId === category.id ? '' : category.id"
	                      class="inline-flex h-10 shrink-0 items-center rounded-xl border border-slate-200 bg-white px-3 text-xs font-bold text-slate-500 transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600"
	                  >
	                    {{ openIconPickerCategoryId === category.id ? '收起' : '选择' }}
	                  </button>
	                </div>
	                <div
	                    v-if="openIconPickerCategoryId === category.id"
	                    class="mt-3 grid max-h-28 grid-cols-8 gap-1.5 overflow-y-auto pr-1 custom-scrollbar"
	                >
	                  <button
	                      v-for="icon in iconOptions"
	                      :key="icon"
                      type="button"
                      @click="selectCategoryIcon(category, icon)"
                      class="flex h-8 items-center justify-center rounded-xl text-base transition-all hover:bg-blue-50 hover:scale-105"
                      :class="category.iconName === icon ? 'bg-blue-600 text-white shadow-sm shadow-blue-200' : 'bg-slate-50 text-slate-600'"
                  >
                    {{ icon }}
                  </button>
                </div>
              </div>

              <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">名称</label>
              <input v-model="category.name" class="h-10 w-full rounded-2xl border border-slate-200 bg-white px-3 text-sm font-bold text-slate-800 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" />

              <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">ID</label>
              <input v-model="category.id" class="h-10 w-full rounded-2xl border border-slate-200 bg-white px-3 font-mono text-xs text-slate-600 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" />
            </div>

            <div class="mt-4 flex justify-end">
              <button
                  type="button"
                  @click="deleteCategory(category.id)"
                  class="inline-flex h-9 items-center gap-1.5 rounded-2xl border border-red-100 bg-white px-3 text-xs font-bold text-red-500 transition-all hover:bg-red-50"
              >
                <Trash2 class="h-3.5 w-3.5" />
                删除分类
              </button>
            </div>
          </div>
        </div>
      </section>

      <section class="rounded-[2rem] border border-slate-200 bg-white p-5 shadow-sm">
        <div class="mb-5 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
          <div>
            <h2 class="text-lg font-black text-slate-900">默认书签</h2>
            <p class="mt-1 text-xs text-slate-400">这些书签会按当前顺序写入模板，新用户注册后会复制一份到自己的账户。</p>
          </div>
          <button
              type="button"
              @click="addSite()"
              class="inline-flex h-10 items-center justify-center gap-2 rounded-2xl bg-blue-600 px-4 text-xs font-bold text-white shadow-sm shadow-blue-200 transition-all hover:-translate-y-0.5 hover:bg-blue-700"
          >
            <Plus class="h-3.5 w-3.5" />
            新增书签
          </button>
        </div>

        <div class="grid gap-4 xl:grid-cols-2">
          <div
              v-for="(site, index) in sites"
              :key="site.id"
              :data-seed-site-id="site.id"
              class="rounded-3xl border p-4 transition-all duration-300 hover:border-slate-200 hover:bg-white hover:shadow-sm"
              :class="highlightedSiteId === site.id ? 'border-blue-300 bg-blue-50/70 shadow-lg shadow-blue-100/70 ring-4 ring-blue-500/10' : 'border-slate-100 bg-slate-50/80'"
          >
            <div class="mb-4 flex items-start justify-between gap-3">
              <div class="flex min-w-0 items-center gap-3">
                <div class="flex h-11 w-11 shrink-0 items-center justify-center overflow-hidden rounded-2xl bg-white shadow-sm ring-1 ring-slate-100">
                  <img v-if="site.icon" :src="site.icon" class="h-5 w-5 object-contain" />
                  <span v-else class="text-xs font-black text-slate-400">{{ site.name?.charAt(0) || 'L' }}</span>
                </div>
                <div class="min-w-0">
                  <p class="truncate text-sm font-black text-slate-800">{{ site.name || '未命名书签' }}</p>
                  <p class="mt-0.5 truncate text-xs text-slate-400">{{ site.url }}</p>
                </div>
              </div>
              <div class="flex shrink-0 items-center gap-1.5">
                <button
                    type="button"
                    :disabled="index === 0"
                    @click="moveSite(index, -1)"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-400 transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600 disabled:cursor-not-allowed disabled:opacity-40"
                    title="上移书签"
                >
                  <ArrowUp class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    :disabled="index === sites.length - 1"
                    @click="moveSite(index, 1)"
                    class="inline-flex h-8 w-8 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-400 transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600 disabled:cursor-not-allowed disabled:opacity-40"
                    title="下移书签"
                >
                  <ArrowDown class="h-3.5 w-3.5" />
                </button>
                <button
                    type="button"
                    @click="deleteSite(site.id)"
                    class="inline-flex h-8 items-center gap-1.5 rounded-xl border border-red-100 bg-white px-2.5 text-xs font-bold text-red-500 transition-all hover:bg-red-50"
                >
                  <Trash2 class="h-3.5 w-3.5" />
                  删除
                </button>
              </div>
            </div>

            <div class="grid gap-3 md:grid-cols-2">
              <div class="space-y-1.5">
                <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">书签名称</label>
                <input v-model="site.name" class="h-10 w-full rounded-2xl border border-slate-200 bg-white px-3 text-sm font-bold text-slate-800 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" />
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">所属分类</label>
                <select v-model="site.categoryId" class="h-10 w-full rounded-2xl border border-slate-200 bg-white px-3 text-sm font-bold text-slate-700 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10">
                  <option v-for="category in categories" :key="category.id" :value="category.id">
                    {{ category.iconName }} {{ category.name }}
                  </option>
                </select>
              </div>
              <div class="space-y-1.5 md:col-span-2">
                <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">URL</label>
                <div class="flex gap-2">
                  <input v-model="site.url" class="h-10 min-w-0 flex-1 rounded-2xl border border-slate-200 bg-white px-3 text-sm text-slate-700 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" />
                  <button
                      type="button"
                      @click="fillSiteMetadata(site)"
                      class="inline-flex h-10 shrink-0 items-center gap-1.5 rounded-2xl border border-slate-200 bg-white px-3 text-xs font-bold text-slate-500 transition-all hover:border-blue-200 hover:text-blue-600"
                  >
                    <Wand2 class="h-3.5 w-3.5" />
                    自动获取
                  </button>
                </div>
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">书签 ID</label>
                <input v-model="site.id" class="h-10 w-full rounded-2xl border border-slate-200 bg-white px-3 font-mono text-xs text-slate-600 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" />
              </div>
              <div class="space-y-1.5">
                <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">图标 URL</label>
                <input v-model="site.icon" class="h-10 w-full rounded-2xl border border-slate-200 bg-white px-3 text-xs text-slate-600 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" />
              </div>
              <div class="space-y-1.5 md:col-span-2">
                <label class="text-[10px] font-bold uppercase tracking-widest text-slate-400">描述</label>
                <textarea v-model="site.description" class="min-h-[72px] w-full resize-none rounded-2xl border border-slate-200 bg-white px-3 py-2 text-sm text-slate-700 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10"></textarea>
              </div>
            </div>

            <p v-if="!categoryIdSet.has(site.categoryId)" class="mt-3 text-xs font-bold text-red-500">
              当前分类引用无效，请重新选择分类。
            </p>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
