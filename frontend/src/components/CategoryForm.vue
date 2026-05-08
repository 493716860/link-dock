<script setup lang="ts">
import { ref, watch } from 'vue';
import type { Category } from '../types';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';

const props = defineProps<{
  isOpen: boolean;
  category: Category | null;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'save', data: { name: string; iconName: string }): void;
}>();

const name = ref('');
const iconName = ref('');

const ICON_POOL = [
  '📁', '🛠️', '🎨', '💻', '☕', '📚', '🚀', '⭐',
  '🔥', '💡', '🎵', '🎮', '🎬', '🖼️', '📦', '📎',
  '📊', '⚙️', '🔒', '☁️', '🌍', '🧭', '📰', '🛒',
  '💰', '🏥', '🎓', '🎪', '⚽', '💬', '📝', '📌',
  '🧰', '✨', '👨‍💻', '🧪', '🖥️', '📱', '📈', '🎯'
];

watch(() => props.isOpen, (val) => {
  if (val) {
    name.value = props.category?.name || '';
    iconName.value = props.category?.iconName || '📁';
  }
});

const handleSave = () => {
  if (!name.value.trim() || !iconName.value) return;
  emit('save', { name: name.value.trim(), iconName: iconName.value });
};

const selectIcon = (icon: string) => {
  iconName.value = icon;
};

const handleOpenChange = (open: boolean) => {
  if (!open) emit('close');
};
</script>

<template>
  <Dialog :open="isOpen" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[360px] p-0 overflow-hidden rounded-[20px] border border-white/60 shadow-[0_32px_64px_-12px_rgba(0,0,0,0.15)] bg-white/95 backdrop-blur-3xl ring-0">
      <DialogHeader class="p-6 pb-3">
        <DialogTitle class="text-xl font-black text-slate-800 tracking-tight">{{ category ? '编辑分类' : '新建分类' }}</DialogTitle>
        <DialogDescription class="text-xs text-slate-500 font-medium mt-1">
          为你的分类选择一个好记的名字和图标。
        </DialogDescription>
      </DialogHeader>

      <div class="px-6 py-1 space-y-5">
        <div class="space-y-1.5">
          <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">分类名称</label>
          <input
              v-model="name"
              placeholder="例如：设计资源"
              class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              autocomplete="off"
          />
        </div>

        <div class="space-y-2">
          <div class="flex items-center justify-between mb-1">
            <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">选择图标</label>
            <div class="text-lg bg-white shadow-sm border border-slate-100 w-8 h-8 flex items-center justify-center rounded-lg transition-transform hover:scale-110">{{ iconName }}</div>
          </div>
          <div class="grid grid-cols-6 gap-2 p-2.5 border border-slate-200/60 bg-slate-100/50 rounded-lg max-h-[140px] overflow-y-auto custom-scrollbar">
            <button
                v-for="icon in ICON_POOL"
                :key="icon"
                type="button"
                class="w-7 h-7 flex items-center justify-center transition-all text-base rounded hover:bg-white hover:shadow-sm"
                :class="iconName === icon ? 'bg-white shadow-sm ring-1 ring-blue-500/30 scale-110' : 'opacity-70 hover:opacity-100'"
                @click="selectIcon(icon)"
            >
              {{ icon }}
            </button>
          </div>
        </div>
      </div>

      <div class="p-6 pt-4 flex justify-end gap-2.5 border-t border-slate-100 mt-2">
        <button
            class="px-5 h-9 rounded-lg text-sm font-bold text-slate-500 hover:bg-slate-100 transition-all active:scale-95"
            @click="emit('close')"
        >
          取消
        </button>
        <button
            class="px-6 h-9 rounded-lg bg-blue-600 text-white font-bold text-sm shadow-md shadow-blue-200/50 hover:bg-blue-700 hover:-translate-y-px transition-all active:scale-[0.97]"
            @click="handleSave"
        >
          保存
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