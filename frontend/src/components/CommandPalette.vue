<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import Fuse from 'fuse.js';
import { Dialog, DialogContent } from '@/components/ui/dialog';
import { Search, CornerDownLeft, Hash, ArrowRight, SearchX } from 'lucide-vue-next';
import type { Site, Category } from '@/types';

const props = defineProps<{
  isOpen: boolean;
  sites: Site[];
  categories: Category[];
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'selectSite', site: Site): void;
  (e: 'selectCategory', categoryId: string | null): void;
}>();

const query = ref('');
const selectedIndex = ref(0);
const inputRef = ref<HTMLInputElement | null>(null);

const fuse = computed(() => {
  const items = [
    ...props.categories.map(c => ({ ...c, type: 'category' })),
    ...props.sites.map(s => ({ ...s, type: 'site' }))
  ];
  return new Fuse(items, {
    keys: [
      { name: 'name', weight: 0.7 },
      { name: 'description', weight: 0.3 },
      { name: 'url', weight: 0.2 },
      { name: 'tagsText', weight: 0.25 }
    ],
    threshold: 0.4,
  });
});

const results = computed(() => {
  if (!query.value) {
    return [
      ...props.categories.slice(0, 3).map(c => ({ item: { ...c, type: 'category' } })),
      ...[...props.sites]
          .sort((a, b) => {
            if ((b.isFavorite ? 1 : 0) !== (a.isFavorite ? 1 : 0)) {
              return (b.isFavorite ? 1 : 0) - (a.isFavorite ? 1 : 0);
            }
            return new Date(b.lastVisitedAt || b.createdAt || 0).getTime() - new Date(a.lastVisitedAt || a.createdAt || 0).getTime();
          })
          .slice(0, 5)
          .map(s => ({ item: { ...s, type: 'site' } }))
    ];
  }
  return fuse.value.search(query.value).slice(0, 8);
});

watch(results, () => {
  selectedIndex.value = 0;
});

const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'ArrowDown') {
    e.preventDefault();
    selectedIndex.value = (selectedIndex.value + 1) % results.value.length;
  } else if (e.key === 'ArrowUp') {
    e.preventDefault();
    selectedIndex.value = (selectedIndex.value - 1 + results.value.length) % results.value.length;
  } else if (e.key === 'Enter') {
    e.preventDefault();
    const selected = results.value[selectedIndex.value]?.item;
    if (selected) {
      handleSelect(selected);
    }
  }
};

const handleSelect = (item: any) => {
  if (item.type === 'site') {
    emit('selectSite', item);
  } else {
    emit('selectCategory', String(item.id));
  }
  query.value = '';
  emit('close');
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    setTimeout(() => inputRef.value?.focus(), 100);
  }
});
</script>

<template>
  <Dialog :open="isOpen" @update:open="emit('close')">
    <DialogContent class="max-w-xl p-0 overflow-hidden rounded-[24px] border border-white/60 shadow-[0_40px_80px_-16px_rgba(0,0,0,0.2)] bg-white/95 backdrop-blur-3xl ring-0">

      <div class="relative flex items-center px-5 py-4 border-b border-slate-200/60 bg-white/50">
        <Search class="w-5 h-5 mr-3 text-blue-500" />
        <input
            ref="inputRef"
            v-model="query"
            @keydown="handleKeyDown"
            class="w-full bg-transparent border-none outline-none text-base font-bold placeholder:text-slate-300 text-slate-800"
            placeholder="搜索你想找的内容、链接或分类..."
        />
        <div class="flex items-center gap-1.5 px-2 py-1 bg-white border border-slate-200 rounded-md text-[10px] font-black text-slate-400 shadow-sm">
          ESC
        </div>
      </div>

      <div class="max-h-[50vh] overflow-y-auto p-2 custom-scrollbar">
        <div v-if="results.length > 0">
          <div
              v-for="(result, index) in results"
              :key="result.item.id + '-' + result.item.type"
              @mouseenter="selectedIndex = index"
              @click="handleSelect(result.item)"
              class="group flex cursor-pointer items-center gap-4 rounded-xl px-4 py-3 transition-all duration-200"
              :class="index === selectedIndex ? 'bg-blue-50 text-blue-900 shadow-sm border border-blue-100/50' : 'text-slate-600 hover:bg-slate-50/80 border border-transparent'"
          >
            <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-slate-100 bg-white shadow-sm transition-transform group-hover:scale-105">
              <span v-if="result.item.type === 'category'" class="text-base">{{ result.item.iconName || '📁' }}</span>
              <img v-else-if="result.item.icon" :src="result.item.icon" class="h-5 w-5 object-contain" />
              <Hash v-else class="h-4 w-4 text-slate-400" />
            </div>

            <div class="flex-1 min-w-0 text-left">
              <div class="flex items-center gap-2">
                <span class="text-sm font-bold truncate transition-colors" :class="index === selectedIndex ? 'text-blue-800' : 'text-slate-800'">
                  {{ result.item.name }}
                </span>
                <span v-if="result.item.type === 'category'" class="text-[9px] px-2 py-0.5 bg-blue-100 text-blue-600 rounded-full font-black uppercase tracking-widest border border-blue-200/50">分类目录</span>
                <span v-else-if="result.item.isFavorite" class="text-[9px] px-2 py-0.5 bg-amber-100 text-amber-600 rounded-full font-black uppercase tracking-widest border border-amber-200/50">常用</span>
              </div>
              <span class="text-[11px] text-slate-400 truncate block mt-0.5 font-medium" :class="index === selectedIndex ? 'text-blue-500/80' : ''">
                {{ result.item.type === 'site' ? result.item.url : (result.item.description || '快速导航到此分类') }}
              </span>
              <span v-if="result.item.type === 'site' && result.item.tagsText" class="text-[10px] text-slate-400 truncate block mt-1">
                #{{ result.item.tagsText }}
              </span>
            </div>

            <div v-if="index === selectedIndex" class="flex items-center gap-1.5 shrink-0 px-2 py-1 bg-white border border-blue-100 shadow-sm rounded-lg text-[10px] font-black text-blue-500">
              <CornerDownLeft class="w-3 h-3" />
            </div>
          </div>
        </div>
        <div v-else class="py-16 flex flex-col items-center justify-center text-slate-300">
          <SearchX class="w-10 h-10 mb-4 opacity-20" />
          <p class="text-xs font-bold uppercase tracking-widest">没有找到相关结果</p>
        </div>
      </div>

      <div class="px-5 py-3 border-t border-slate-200/60 bg-slate-50/50 flex items-center justify-between text-[10px] font-bold text-slate-400 uppercase tracking-widest">
        <div class="flex gap-6">
          <span class="flex items-center gap-1.5"><ArrowRight class="w-3.5 h-3.5" /> 切换选择</span>
          <span class="flex items-center gap-1.5"><CornerDownLeft class="w-3.5 h-3.5" /> 确认跳转</span>
        </div>
        <span class="opacity-50">全局检索 Spotlight</span>
      </div>
    </DialogContent>
  </Dialog>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar { width: 6px; }
.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
.custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(203, 213, 225, 0.6); border-radius: 10px; }
</style>
