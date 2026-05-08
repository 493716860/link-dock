<script setup lang="ts">
import { computed, ref } from 'vue';
import draggable from 'vuedraggable';
import type { Category, Site } from '../types';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Plus, GripVertical, MoreHorizontal, LayoutGrid } from 'lucide-vue-next';

const props = defineProps<{
  categories: Category[];
  sites: Site[];
  activeCategoryId: string | null;
  canEdit: boolean;
  draggedSiteId?: string | null;
  collapsed?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select-category', id: string | null): void;
  (e: 'add-site'): void;
  (e: 'add-category'): void;
  (e: 'edit-category', category: Category): void;
  (e: 'delete-category', id: string): void;
  (e: 'update-categories-order', ids: string[]): void;
  (e: 'move-site-to-category', payload: { siteId: string; categoryId: string }): void;
}>();

const dropTargetCategoryId = ref<string | null>(null);

const allSitesCount = computed(() => props.sites.length);

const categorySiteCounts = computed(() => {
  const counts: Record<string, number> = {};
  props.sites.forEach(site => { counts[site.categoryId] = (counts[site.categoryId] || 0) + 1; });
  return counts;
});

const draggableCategories = computed({
  get: () => props.categories,
  set: (val) => emit('update-categories-order', val.map(c => c.id))
});

const handleCategoryDragOver = (event: DragEvent, categoryId: string) => {
  if (!props.draggedSiteId) return;
  event.preventDefault();
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'move';
  }
  dropTargetCategoryId.value = categoryId;
};

const handleCategoryDragLeave = (categoryId: string) => {
  if (dropTargetCategoryId.value === categoryId) {
    dropTargetCategoryId.value = null;
  }
};

const handleCategoryDrop = (event: DragEvent, categoryId: string) => {
  if (!props.draggedSiteId) return;
  event.preventDefault();
  dropTargetCategoryId.value = null;
  emit('move-site-to-category', {
    siteId: props.draggedSiteId,
    categoryId,
  });
};
</script>

<template>
  <div class="flex h-full flex-col border-transparent transition-all duration-300 ease-out" :class="collapsed ? 'px-2 py-4' : 'px-4 py-6'">
    <div class="flex h-full flex-col transition-all duration-300 ease-out" :class="collapsed ? 'space-y-3' : 'space-y-4'">
      <div v-if="canEdit" class="flex shrink-0 transition-all duration-300" :class="collapsed ? 'justify-center' : 'justify-end px-1'">
          <button
              class="rounded-md hover:bg-white hover:shadow-sm flex items-center justify-center text-slate-400 hover:text-blue-600 transition-all active:scale-90"
              :class="collapsed ? 'h-10 w-10 rounded-2xl border border-slate-100 bg-white/80 shadow-sm' : 'w-6 h-6'"
              title="新增分类"
              aria-label="新增分类"
              @click="emit('add-category')"
          >
            <Plus class="h-4 w-4" />
          </button>
      </div>

      <div class="flex-1 overflow-y-auto custom-scrollbar" :class="collapsed ? 'px-0' : 'pr-1'">
        <div class="space-y-1 pb-4">
          <button
              class="relative w-full flex items-center rounded-2xl text-sm font-bold transition-all duration-300 border"
              :class="activeCategoryId === null ? 'bg-white shadow-md text-blue-600 border-white' : 'text-slate-500 hover:bg-white/60 border-transparent'"
              :title="collapsed ? '全部' : undefined"
              @click="emit('select-category', null)"
          >
            <div class="flex items-center min-w-0 transition-all duration-300" :class="collapsed ? 'h-12 w-full justify-center' : 'gap-2.5 flex-1 py-2 px-3'">
              <div class="shrink-0 rounded-lg bg-white border border-slate-100 shadow-sm flex items-center justify-center transition-all duration-300"
                   :class="collapsed ? 'h-9 w-9 rounded-2xl' : 'w-6 h-6'">
                <LayoutGrid class="shrink-0 transition-all duration-300" :class="collapsed ? 'h-5 w-5' : 'h-3.5 w-3.5'" />
              </div>
              <span v-if="!collapsed" class="truncate">全部</span>
            </div>
            <span
                v-if="!collapsed"
                class="font-black rounded-full shrink-0 transition-all duration-300"
                :class="[
                  activeCategoryId === null ? 'bg-blue-50 text-blue-600' : 'bg-white text-slate-400',
                  'text-[10px] px-2 py-0.5 ml-auto mr-3'
                ]"
            >
              {{ allSitesCount }}
            </span>
          </button>

          <draggable
              v-model="draggableCategories"
              item-key="id"
              class="space-y-1"
              handle=".drag-handle"
              :disabled="!canEdit || collapsed"
              ghost-class="opacity-40"
          >
            <template #item="{ element: category }">
              <div class="group relative flex w-full items-center">
                <div
                    @click="emit('select-category', category.id)"
                    @dragover="handleCategoryDragOver($event, category.id)"
                    @dragleave="handleCategoryDragLeave(category.id)"
                    @drop="handleCategoryDrop($event, category.id)"
                    class="relative w-full rounded-2xl text-sm font-bold transition-all duration-300 flex items-center cursor-pointer border"
                    :class="[
                      dropTargetCategoryId === category.id
                        ? 'bg-blue-50 shadow-md text-blue-600 border-blue-200 ring-2 ring-blue-200/60'
                        : activeCategoryId === category.id
                          ? 'bg-white shadow-md text-blue-600 border-white'
                          : 'text-slate-500 hover:bg-white/60 border-transparent',
                      collapsed ? 'h-12 justify-center px-0 py-0' : 'py-2 px-3'
                    ]"
                    :title="collapsed ? category.name : undefined"
                >
                  <div class="flex items-center min-w-0 transition-all duration-300" :class="collapsed ? 'justify-center' : 'gap-2.5 flex-1'">
                    <div v-if="canEdit && !collapsed" class="drag-handle opacity-0 group-hover:opacity-100 transition-opacity cursor-grab shrink-0 -ml-1">
                      <GripVertical class="h-3.5 w-3.5 text-slate-300" />
                    </div>
                    <div class="shrink-0 rounded-lg bg-white border border-slate-100 shadow-sm flex items-center justify-center transition-all duration-300 group-hover:scale-105"
                         :class="collapsed ? 'h-9 w-9 rounded-2xl' : 'w-6 h-6'">
                      <span class="leading-none" :class="collapsed ? 'text-base' : 'text-[11px]'">{{ category.iconName || '📁' }}</span>
                    </div>
                    <span v-if="!collapsed" class="truncate">{{ category.name }}</span>
                  </div>

	                  <div v-if="!collapsed" class="relative flex items-center justify-center shrink-0 ml-auto w-6 h-6">
	                    <span class="flex items-center justify-center font-black transition-all duration-200 rounded-full"
	                          :class="[
	                            activeCategoryId === category.id ? 'bg-blue-50 text-blue-600' : 'bg-white text-slate-400',
	                            'absolute text-[10px] group-hover:opacity-0 group-hover:scale-75 group-has-[[data-state=open]]:opacity-0 group-has-[[data-state=open]]:scale-75 px-1.5 py-0.5'
	                          ]">
	                      {{ categorySiteCounts[category.id] || 0 }}
	                    </span>

	                    <div v-if="canEdit" class="absolute inset-0 flex items-center justify-center">
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <button @click.stop class="h-6 w-6 rounded-md hover:bg-slate-100 flex items-center justify-center text-slate-400 opacity-0 group-hover:opacity-100 data-[state=open]:opacity-100 transition-all duration-200">
                            <MoreHorizontal class="h-4 w-4" />
                          </button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent side="bottom" align="end" :side-offset="4" class="min-w-[72px] w-[72px] rounded-xl border-slate-100 p-1 shadow-lg bg-white/95 backdrop-blur-xl z-[60]">
                          <DropdownMenuItem @click.stop="emit('edit-category', category)" class="flex items-center justify-center rounded-lg py-1.5 text-[11px] font-bold text-slate-600 hover:bg-slate-50 cursor-pointer">
                            编辑
                          </DropdownMenuItem>
                          <DropdownMenuSeparator class="my-0.5 border-slate-100" />
                          <DropdownMenuItem @click.stop="emit('delete-category', category.id)" class="flex items-center justify-center rounded-lg py-1.5 text-[11px] font-bold text-red-500 hover:bg-red-50 cursor-pointer">
                            删除
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  </div>

                </div>
              </div>
            </template>
          </draggable>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 4px;
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
</style>
