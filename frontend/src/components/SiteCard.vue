<script setup lang="ts">
import { ref, computed } from 'vue';
import type { Site } from '../types';
import { Edit2, Trash2, Star, Clock3, Inbox } from 'lucide-vue-next';

const props = defineProps<{
  site: Site;
  canEdit: boolean;
  compact?: boolean;
}>();

const emit = defineEmits<{
  (e: 'edit', site: Site): void;
  (e: 'delete', id: string): void;
  (e: 'open', site: Site): void;
  (e: 'toggle-favorite', site: Site): void;
  (e: 'cycle-status', site: Site): void;
}>();

const iconLoadError = ref(false);

const openSite = () => {
  emit('open', props.site);
}

const displayUrl = computed(() => {
  try {
    return new URL(props.site.url).hostname;
  } catch {
    return props.site.url;
  }
});

const workflowLabel = computed(() => {
  if (props.site.workflowStatus === 'read_later') return '稍后读';
  if (props.site.workflowStatus === 'unorganized') return '未整理';
  return '';
});

const tags = computed(() => {
  return (props.site.tagsText || '')
      .split(',')
      .map(tag => tag.trim())
      .filter(Boolean)
      .slice(0, props.compact ? 2 : 3);
});
</script>

<template>
  <div
      v-if="compact"
      :data-site-id="site.id"
      class="group cursor-pointer rounded-2xl border border-slate-200 bg-white px-3 py-3 shadow-sm transition-all duration-200 hover:-translate-y-0.5 hover:border-blue-200 hover:shadow-md active:scale-[0.99]"
      @click="openSite"
  >
    <div class="flex items-start gap-3">
      <div class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-xl border border-slate-200 bg-slate-50 shadow-[inset_0_1px_0_rgba(255,255,255,0.9)]">
        <img v-if="site.icon && !iconLoadError" :src="site.icon" class="h-5 w-5 object-contain" referrerpolicy="no-referrer" @error="iconLoadError = true" />
        <div v-else class="flex h-full w-full items-center justify-center bg-slate-100 text-xs font-black text-slate-500">
          {{ site.name.charAt(0).toUpperCase() }}
        </div>
      </div>

      <div class="min-w-0 flex-1">
        <div class="flex items-start justify-between gap-3">
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-1.5">
              <h3 class="truncate text-sm font-bold text-slate-800 transition-colors group-hover:text-blue-600">
                {{ site.name }}
              </h3>
            </div>
            <p class="mt-0.5 line-clamp-2 text-[11px] leading-tight text-slate-500">
              {{ site.description || '暂无描述信息' }}
            </p>
          </div>

        </div>
      </div>
    </div>

    <div class="mt-2 flex flex-wrap items-center gap-1.5">
      <span v-if="workflowLabel" class="inline-flex shrink-0 whitespace-nowrap rounded-full bg-blue-50 px-2 py-1 text-[10px] font-bold text-blue-600">
        {{ workflowLabel }}
      </span>
      <span v-for="tag in tags" :key="tag" class="inline-flex shrink-0 whitespace-nowrap rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-500">
        #{{ tag }}
      </span>
    </div>

    <div class="mt-3 space-y-2 border-t border-slate-100 pt-2.5">
      <div class="flex items-center gap-2">
        <span class="inline-flex min-w-0 max-w-full flex-1 items-center rounded-full bg-slate-100 px-2 py-1 text-[10px] font-semibold text-slate-500">
          <span class="truncate">{{ displayUrl }}</span>
        </span>
        <span v-if="site.visitCount" class="shrink-0 rounded-full bg-slate-50 px-2 py-1 text-[10px] font-semibold text-slate-400">
          {{ site.visitCount }} 次
        </span>
      </div>

      <div v-if="canEdit" class="flex items-center justify-end gap-1 opacity-100 transition-opacity sm:opacity-70 sm:group-hover:opacity-100">
        <button
            @click.stop="emit('toggle-favorite', site)"
            class="flex h-7 min-w-7 items-center justify-center rounded-lg border px-1.5 transition-all duration-200"
            :class="site.isFavorite ? 'border-amber-200 bg-amber-50 text-amber-500' : 'border-slate-200 text-slate-400 hover:border-amber-100 hover:bg-amber-50 hover:text-amber-500'"
            title="切换常用"
        >
            <Star class="h-3.5 w-3.5" :class="site.isFavorite ? 'fill-amber-400 text-amber-400' : ''" />
        </button>
        <button
            @click.stop="emit('cycle-status', site)"
            class="flex h-7 w-7 items-center justify-center rounded-lg border border-slate-200 text-slate-400 transition-all duration-200 hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600"
            title="切换状态"
        >
            <Clock3 v-if="site.workflowStatus === 'read_later'" class="h-3.5 w-3.5" />
            <Inbox v-else class="h-3.5 w-3.5" />
        </button>
        <button
            @click.stop="emit('edit', site)"
            class="flex h-7 w-7 items-center justify-center rounded-lg border border-slate-200 text-slate-400 transition-all duration-200 hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600"
            title="编辑书签"
        >
            <Edit2 class="h-3.5 w-3.5" />
        </button>
        <button
            @click.stop="emit('delete', site.id)"
            class="flex h-7 w-7 items-center justify-center rounded-lg border border-slate-200 text-slate-400 transition-all duration-200 hover:border-red-100 hover:bg-red-50 hover:text-red-500"
            title="删除书签"
        >
            <Trash2 class="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  </div>

  <div
      v-else
      :data-site-id="site.id"
      class="group relative p-3 cursor-pointer overflow-hidden rounded-2xl border border-white/80 bg-white/40 backdrop-blur-xl shadow-sm hover:shadow-[0_16px_32px_rgba(0,0,0,0.06)] hover:bg-white/50 transition-all duration-300 hover:-translate-y-1 active:scale-[0.98] ring-1 ring-slate-900/5 flex flex-col h-full"
      @click="openSite"
  >
    <!-- 修改为科技蓝色的常驻渐变 -->
    <div class="absolute inset-0 z-0 opacity-20 group-hover:opacity-40 transition-opacity duration-700 overflow-hidden">
      <div class="absolute -top-[50%] -left-[50%] w-[200%] h-[200%] bg-[conic-gradient(from_0deg,transparent_0deg,rgba(59,130,246,0.6)_90deg,rgba(14,165,233,0.6)_180deg,rgba(6,182,212,0.6)_270deg,transparent_360deg)] animate-[spin_10s_linear_infinite] filter blur-[24px]"></div>
    </div>

    <div class="absolute inset-0 z-10 bg-gradient-to-br from-white/70 via-white/10 to-transparent opacity-80 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none"></div>

      <div class="relative z-20 flex gap-2.5 items-center mb-2.5">
      <div class="w-8 h-8 shrink-0 rounded-[8px] shadow-sm flex items-center justify-center bg-white border border-white/80 group-hover:shadow-md group-hover:scale-105 transition-all duration-300 overflow-hidden">
        <img v-if="site.icon && !iconLoadError" :src="site.icon" class="w-4 h-4 object-contain" referrerpolicy="no-referrer" @error="iconLoadError = true" />
        <div v-else class="w-full h-full flex items-center justify-center text-[11px] font-black text-slate-400 bg-slate-50">
          {{ site.name.charAt(0).toUpperCase() }}
        </div>
      </div>

      <div class="relative z-30 flex-1 min-w-0 flex flex-col justify-center">
        <div class="flex items-center gap-1.5">
          <h3 class="text-[13px] font-bold text-slate-800 group-hover:text-blue-600 transition-colors truncate leading-tight">
            {{ site.name }}
          </h3>
        </div>
        <p class="text-[10px] text-slate-500 truncate mt-0.5 font-medium leading-tight">
          {{ site.description || '暂无描述信息' }}
        </p>
        <div class="mt-1.5 flex flex-wrap items-center gap-1">
          <span v-if="workflowLabel" class="rounded-full bg-blue-50 px-2 py-0.5 text-[9px] font-bold text-blue-600">
            {{ workflowLabel }}
          </span>
          <span v-for="tag in tags" :key="tag" class="rounded-full bg-white/80 px-2 py-0.5 text-[9px] font-semibold text-slate-500 border border-white/80">
            #{{ tag }}
          </span>
        </div>
      </div>
    </div>

    <div class="relative z-30 mt-auto space-y-2 border-t border-slate-200/30 pt-2.5">
      <div class="flex items-center gap-1.5">
        <div class="min-w-0 flex-1 overflow-hidden flex items-center px-1.5 py-0.5 rounded-md bg-white/70 border border-white/80 shadow-[inset_0_1px_2px_rgba(255,255,255,0.8)]">
          <span class="text-[9px] text-slate-500 font-bold tracking-tight truncate group-hover:text-blue-600 transition-colors">
            {{ displayUrl }}
          </span>
        </div>
        <span v-if="site.visitCount" class="shrink-0 rounded-md bg-white/70 px-1.5 py-0.5 text-[9px] font-bold text-slate-400 border border-white/80">
          {{ site.visitCount }} 次
        </span>
      </div>

      <div class="flex items-center justify-end gap-1 opacity-70 group-hover:opacity-100 transition-all duration-300">
        <div v-if="canEdit" class="flex items-center gap-0.5">
          <button
              @click.stop="emit('toggle-favorite', site)"
              class="flex h-6 min-w-6 items-center justify-center rounded border px-1 transition-all duration-200"
              :class="site.isFavorite ? 'border-amber-200 bg-amber-50/90 text-amber-500' : 'border-white/80 text-slate-500 hover:border-amber-100 hover:bg-amber-50/80 hover:shadow-sm hover:text-amber-500'"
          >
            <Star class="w-2.5 h-2.5" :class="site.isFavorite ? 'fill-amber-400 text-amber-400' : ''" />
          </button>
          <button
              @click.stop="emit('cycle-status', site)"
              class="h-6 w-6 rounded border border-white/80 hover:border-blue-100 hover:bg-blue-50/80 hover:shadow-sm flex items-center justify-center text-slate-500 hover:text-blue-500 transition-all duration-200"
          >
            <Clock3 v-if="site.workflowStatus === 'read_later'" class="w-2.5 h-2.5" />
            <Inbox v-else class="w-2.5 h-2.5" />
          </button>
          <button
              @click.stop="emit('edit', site)"
              class="h-6 w-6 rounded border border-white/80 hover:border-white hover:bg-white/80 hover:shadow-sm flex items-center justify-center text-slate-500 hover:text-blue-600 transition-all duration-200"
          >
            <Edit2 class="w-2.5 h-2.5" />
          </button>
          <button
              @click.stop="emit('delete', site.id)"
              class="h-6 w-6 rounded border border-white/80 hover:border-red-100 hover:bg-red-50/80 hover:shadow-sm flex items-center justify-center text-slate-500 hover:text-red-500 transition-all duration-200"
          >
            <Trash2 class="w-2.5 h-2.5" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
