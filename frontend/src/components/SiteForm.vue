<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import type { Category, Site, WorkflowStatus } from '../types';
import { api } from '../api';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Loader2, Wand2, AlertTriangle, Sparkles, FolderSearch } from 'lucide-vue-next';

const props = defineProps<{
  isOpen: boolean;
  site: Site | null;
  categories: Category[];
  sites: Site[];
  initialCategoryId: string | null;
  initialData: { url?: string, name?: string, description?: string } | null;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', data: Omit<Site, 'id' | 'createdAt' | 'updatedAt'>): void;
  (e: 'edit-existing', site: Site): void;
}>();

const name = ref('');
const url = ref('');
const description = ref('');
const categoryId = ref('');
const icon = ref('');
const tagsText = ref('');
const workflowStatus = ref<WorkflowStatus>('unorganized');
const isFetching = ref(false);
const hasManuallySelectedCategory = ref(false);
const iconLoadError = ref(false);
const initialNormalizedUrl = ref('');
const initialIcon = ref('');

const normalizeUrl = (raw: string) => {
  const target = raw.trim();
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

const currentNormalizedUrl = computed(() => normalizeUrl(url.value));

const isImageIcon = computed(() => /^https?:\/\//i.test(icon.value.trim()));

const getGeneratedIcon = () => {
  const source = name.value.trim() || currentHostname.value || url.value.trim();
  const firstChar = Array.from(source).find(char => /\S/.test(char));
  if (!firstChar) return '#';
  return /[a-z]/i.test(firstChar) ? firstChar.toUpperCase() : firstChar;
};

const duplicateSite = computed(() => {
  if (!currentNormalizedUrl.value) return null;
  return props.sites.find(existing => {
    if (props.site?.id && existing.id === props.site.id) return false;
    return normalizeUrl(existing.url) === currentNormalizedUrl.value;
  }) || null;
});

const currentHostname = computed(() => {
  if (!currentNormalizedUrl.value) return '';
  try {
    return new URL(currentNormalizedUrl.value).hostname;
  } catch {
    return '';
  }
});

const sameDomainSites = computed(() => {
  if (!currentHostname.value) return [];
  return props.sites.filter(existing => {
    if (duplicateSite.value?.id === existing.id) return false;
    if (props.site?.id && existing.id === props.site.id) return false;
    try {
      return new URL(normalizeUrl(existing.url)).hostname === currentHostname.value;
    } catch {
      return false;
    }
  });
});

const recommendedCategoryId = computed(() => {
  if (sameDomainSites.value.length === 0) return null;
  const categoryCount = new Map<string, number>();
  sameDomainSites.value.forEach(existing => {
    categoryCount.set(existing.categoryId, (categoryCount.get(existing.categoryId) || 0) + 1);
  });
  const best = [...categoryCount.entries()].sort((a, b) => b[1] - a[1])[0];
  return best?.[0] || null;
});

const recommendedCategory = computed(() => {
  return props.categories.find(cat => String(cat.id) === String(recommendedCategoryId.value)) || null;
});

const recentCategoryId = computed(() => {
  return window.localStorage.getItem('linkdock-recent-category-id');
});

const previewIcon = computed(() => icon.value.trim() || getGeneratedIcon());

watch(() => icon.value, () => {
  iconLoadError.value = false;
});

watch(currentNormalizedUrl, (nextUrl) => {
  if (!props.site) return;
  if (nextUrl && nextUrl !== initialNormalizedUrl.value && icon.value === initialIcon.value) {
    icon.value = '';
  }
});

watch(() => props.isOpen, (val) => {
  if (val) {
    name.value = props.site?.name || props.initialData?.name || '';
    url.value = props.site?.url || props.initialData?.url || '';
    description.value = props.site?.description || props.initialData?.description || '';
    categoryId.value = props.site?.categoryId || props.initialCategoryId || (props.categories.length > 0 ? props.categories[0].id : '');
    icon.value = props.site?.icon || '';
    initialIcon.value = props.site?.icon || '';
    initialNormalizedUrl.value = normalizeUrl(props.site?.url || props.initialData?.url || '');
    tagsText.value = props.site?.tagsText || '';
    workflowStatus.value = props.site?.workflowStatus || 'unorganized';
    isFetching.value = false;
    iconLoadError.value = false;
    hasManuallySelectedCategory.value = !!props.site;

    if (props.initialData?.url && !props.site) {
      handleUrlBlur();
    }
  }
});

watch(() => recommendedCategoryId.value, (nextCategoryId) => {
  if (props.site || hasManuallySelectedCategory.value || !nextCategoryId) return;
  if (!categoryId.value || categoryId.value === recentCategoryId.value) {
    categoryId.value = nextCategoryId;
  }
});

const fetchSiteMetadata = async (overwrite = false) => {
  const targetUrl = url.value.trim();
  if (!targetUrl) return;
  isFetching.value = true;
  try {
    const meta = await api.fetchUrlMetadata(targetUrl);
    if (meta.success) {
      if (overwrite || !name.value) name.value = meta.title || name.value;
      if (overwrite || !description.value) description.value = meta.description || description.value;
      if (overwrite || !icon.value) icon.value = meta.icon || '';
    }
  } catch (error) {
    console.error("Fetch metadata error:", error);
  } finally {
    isFetching.value = false;
  }
};

const handleUrlBlur = async () => {
  if (props.site) return;
  await fetchSiteMetadata(false);
};

const handleSave = () => {
  if (!name.value.trim() || !url.value.trim() || !categoryId.value || duplicateSite.value) return;
  emit('save', {
    name: name.value.trim(),
    url: url.value.trim(),
    description: description.value.trim(),
    categoryId: categoryId.value,
    icon: icon.value.trim() || getGeneratedIcon(),
    tagsText: tagsText.value.trim(),
    workflowStatus: workflowStatus.value,
    isFavorite: props.site?.isFavorite || false,
    isPublic: true
  });
};
</script>

<template>
  <Dialog :open="isOpen" @update:open="(open) => { if (!open) emit('close') }">
    <DialogContent class="sm:max-w-[400px] p-0 overflow-hidden rounded-[20px] border border-white/60 shadow-[0_32px_64px_-12px_rgba(0,0,0,0.15)] bg-white/95 backdrop-blur-3xl ring-0">
      <DialogHeader class="px-6 pt-6 pb-3">
        <DialogTitle class="text-xl font-black text-slate-800 tracking-tight">
          {{ site ? '编辑书签' : '添加新书签' }}
        </DialogTitle>
        <DialogDescription class="text-xs text-slate-500 font-medium mt-1">
          输入 URL 后可尝试自动获取网站信息，并给出归类建议。
        </DialogDescription>
      </DialogHeader>

      <div class="px-6 py-1 max-h-[60vh] overflow-y-auto space-y-4 custom-scrollbar">
        <div class="flex items-center gap-4 rounded-2xl border border-slate-200 bg-gradient-to-br from-white to-slate-50 px-4 py-3 shadow-sm">
          <div class="flex h-14 w-14 shrink-0 items-center justify-center overflow-hidden rounded-2xl border border-white bg-white text-2xl font-black text-slate-700 shadow-md shadow-slate-200/70">
            <img
                v-if="isImageIcon && !iconLoadError"
                :src="icon"
                class="h-8 w-8 object-contain"
                referrerpolicy="no-referrer"
                @error="iconLoadError = true"
            />
            <span v-else class="leading-none">{{ previewIcon }}</span>
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-black text-slate-800">{{ name || '等待获取网站信息' }}</p>
            <p class="mt-1 truncate text-xs font-medium text-slate-400">{{ currentHostname || '输入网址后自动识别 logo、名称和描述' }}</p>
          </div>
        </div>

        <div v-if="duplicateSite" class="rounded-2xl border border-amber-200 bg-amber-50 px-4 py-3">
          <div class="flex items-start gap-3">
            <AlertTriangle class="mt-0.5 h-4 w-4 shrink-0 text-amber-500" />
            <div class="min-w-0 flex-1">
              <p class="text-sm font-bold text-amber-700">这个网址已经保存过了</p>
              <p class="mt-1 text-xs text-amber-700/80 truncate">
                「{{ duplicateSite.name }}」已存在于「{{ categories.find(cat => String(cat.id) === String(duplicateSite.categoryId))?.name || '未知分类' }}」中。
              </p>
              <p class="mt-1 truncate text-[11px] text-amber-700/70">{{ duplicateSite.url }}</p>
              <button
                  type="button"
                  @click="emit('edit-existing', duplicateSite)"
                  class="mt-2 inline-flex h-8 items-center rounded-xl bg-white px-3 text-xs font-bold text-amber-700 shadow-sm"
              >
                编辑已有书签
              </button>
            </div>
          </div>
        </div>

        <div v-else-if="sameDomainSites.length > 0 || recommendedCategory" class="rounded-2xl border border-blue-100 bg-blue-50/70 px-4 py-3">
          <div class="flex items-start gap-3">
            <Sparkles class="mt-0.5 h-4 w-4 shrink-0 text-blue-500" />
            <div class="min-w-0 flex-1 space-y-1.5">
              <p v-if="sameDomainSites.length > 0" class="text-xs font-bold text-blue-700">
                同域名下已收藏 {{ sameDomainSites.length }} 个书签
              </p>
              <p v-if="recommendedCategory" class="flex items-center gap-1 text-xs text-blue-700/90">
                <FolderSearch class="h-3.5 w-3.5" />
                推荐分类：{{ recommendedCategory.iconName }} {{ recommendedCategory.name }}
              </p>
            </div>
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">网站 URL</label>
          <div class="relative">
            <input
                id="siteUrl"
                v-model="url"
                @blur="handleUrlBlur"
                placeholder="https://example.com"
                autocomplete="off"
                :disabled="isFetching"
                class="w-full h-10 px-3 pr-10 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
            />
            <div v-if="isFetching" class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-2">
              <Loader2 class="h-4 w-4 animate-spin text-blue-600" />
            </div>
            <button
                v-else-if="url"
                type="button"
                @click="fetchSiteMetadata(true)"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-blue-600 transition-colors"
                :title="site ? '重新获取网站信息' : '获取网站信息'"
                :aria-label="site ? '重新获取网站信息' : '获取网站信息'"
            >
              <Wand2 class="h-4 w-4" />
            </button>
          </div>
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">网站名称</label>
          <input v-model="name" placeholder="轻舟" autocomplete="off" class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none" />
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">描述信息</label>
          <textarea
              v-model="description"
              placeholder="一句话介绍这个网站..."
              class="w-full px-3 py-2 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none min-h-[70px] resize-none"
          ></textarea>
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">标签</label>
          <input
              v-model="tagsText"
              placeholder="例如：AI, 工作流, 收藏夹"
              autocomplete="off"
              class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
          />
          <p class="text-[11px] text-slate-400 ml-1">多个标签请用英文逗号分隔。</p>
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">所属分类</label>
          <select
              id="siteCat"
              v-model="categoryId"
              @change="hasManuallySelectedCategory = true"
              class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none appearance-none"
          >
            <option v-for="cat in categories" :key="cat.id" :value="cat.id">
              {{ cat.iconName }} {{ cat.name }}
            </option>
          </select>
          <p v-if="recentCategoryId && categoryId === recentCategoryId" class="text-[11px] text-blue-500 ml-1">正在保存到最近使用的分类。</p>
        </div>

        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">书签状态</label>
          <select
              v-model="workflowStatus"
              class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none appearance-none"
          >
            <option value="unorganized">未整理</option>
            <option value="read_later">稍后读</option>
            <option value="none">普通</option>
          </select>
        </div>
      </div>

      <div class="p-6 pt-4 flex justify-end gap-2.5 border-t border-slate-100">
        <button
            class="px-5 h-9 rounded-lg text-sm font-bold text-slate-500 hover:bg-slate-100 transition-all active:scale-95"
            @click="emit('close')"
        >
          取消
        </button>
        <button
            class="px-6 h-9 rounded-lg bg-blue-600 text-white font-bold text-sm shadow-md shadow-blue-200/50 hover:bg-blue-700 hover:-translate-y-px transition-all active:scale-[0.97]"
            @click="handleSave"
            :disabled="isFetching || !!duplicateSite"
        >
          {{ duplicateSite ? '已保存该网址' : props.initialCategoryId === recentCategoryId && !site ? '保存到最近分类' : '保存' }}
        </button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 4px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(203, 213, 225, 0.8); border-radius: 10px; }
</style>
