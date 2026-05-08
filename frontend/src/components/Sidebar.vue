<script setup lang="ts">
import { computed, ref } from 'vue';
import draggable from 'vuedraggable';
import type { Category, Site } from '../types';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Plus, GripVertical, MoreHorizontal, LayoutGrid, Pencil, Trash } from 'lucide-vue-next';

const props = defineProps<{
  categories: Category[];
  sites: Site[];
  activeCategoryId: string | null;
  canEdit: boolean;
  draggedSiteId?: string | null;
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
  <div class="flex h-full flex-col px-4 py-6 border-transparent">
    <div class="space-y-6 flex flex-col h-full">
      <!-- Main Navigation -->
      <div class="space-y-1 shrink-0">
        <button
            class="w-full flex items-center px-4 py-2.5 rounded-2xl text-sm font-bold transition-all duration-300"
            :class="activeCategoryId === null ? 'bg-white shadow-md text-blue-600 border border-white' : 'text-slate-500 hover:bg-white/50 border border-transparent'"
            @click="emit('select-category', null)"
        >
          <div class="flex items-center gap-3 flex-1">
            <LayoutGrid class="h-4 w-4" />
            <span>所有项目</span>
          </div>
          <span class="text-[10px] font-black px-2 py-0.5 rounded-full shrink-0 ml-auto" :class="activeCategoryId === null ? 'bg-blue-50 text-blue-600' : 'bg-white text-slate-400'">{{ allSitesCount }}</span>
        </button>
      </div>

      <!-- Categories Section -->
      <div class="space-y-2 flex-1 flex flex-col min-h-0">
        <div class="flex items-center justify-between px-2 mb-2 shrink-0">
          <h4 class="text-[10px] font-black text-slate-400 uppercase tracking-widest">我的分类</h4>
          <button v-if="canEdit" class="w-6 h-6 rounded-md hover:bg-white hover:shadow-sm flex items-center justify-center text-slate-400 hover:text-blue-600 transition-all active:scale-90" @click="emit('add-category')">
            <Plus class="h-4 w-4" />
          </button>
        </div>

        <div class="flex-1 overflow-y-auto pr-1 custom-scrollbar">
          <draggable
              v-model="draggableCategories"
              item-key="id"
              class="space-y-1 pb-4"
              handle=".drag-handle"
              :disabled="!canEdit"
              ghost-class="opacity-40"
          >
            <template #item="{ element: category }">
              <div class="group relative flex w-full items-center">
                <div
                    @click="emit('select-category', category.id)"
                    @dragover="handleCategoryDragOver($event, category.id)"
                    @dragleave="handleCategoryDragLeave(category.id)"
                    @drop="handleCategoryDrop($event, category.id)"
                    class="w-full py-2 px-3 rounded-2xl text-sm font-bold transition-all duration-300 flex items-center cursor-pointer border"
                    :class="dropTargetCategoryId === category.id
                        ? 'bg-blue-50 shadow-md text-blue-600 border-blue-200 ring-2 ring-blue-200/60'
                        : activeCategoryId === category.id
                          ? 'bg-white shadow-md text-blue-600 border-white'
                          : 'text-slate-500 hover:bg-white/60 border-transparent'"
                >
                  <div class="flex items-center gap-2.5 flex-1 min-w-0">
                    <div v-if="canEdit" class="drag-handle opacity-0 group-hover:opacity-100 transition-opacity cursor-grab shrink-0 -ml-1">
                      <GripVertical class="h-3.5 w-3.5 text-slate-300" />
                    </div>
                    <div class="w-6 h-6 shrink-0 rounded-lg bg-white border border-slate-100 shadow-sm flex items-center justify-center transition-transform group-hover:scale-105">
                      <span class="text-[11px] leading-none">{{ category.iconName || '📁' }}</span>
                    </div>
                    <span class="truncate">{{ category.name }}</span>
                  </div>

                  <div class="relative flex items-center justify-center shrink-0 ml-auto w-6 h-6">
                    <span class="absolute flex items-center justify-center text-[10px] font-black transition-all duration-200 group-hover:opacity-0 group-hover:scale-75 group-has-[[data-state=open]]:opacity-0 group-has-[[data-state=open]]:scale-75 px-1.5 py-0.5 rounded-full"
                          :class="activeCategoryId === category.id ? 'bg-blue-50 text-blue-600' : 'bg-white text-slate-400'">
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
