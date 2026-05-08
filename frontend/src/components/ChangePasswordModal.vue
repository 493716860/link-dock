<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Loader2 } from 'lucide-vue-next';
import { api } from '../api';

const props = defineProps<{
  isOpen: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'success'): void;
}>();

const oldPassword = ref('');
const newPassword = ref('');
const confirmPassword = ref('');
const isLoading = ref(false);
const errorMessage = ref('');
const successMessage = ref('');

const PASSWORD_RULE = '至少 8 位，且同时包含字母和数字';

const canSubmit = computed(() => {
  return !isLoading.value
      && oldPassword.value.trim().length > 0
      && newPassword.value.trim().length > 0
      && confirmPassword.value.trim().length > 0;
});

const resetState = () => {
  oldPassword.value = '';
  newPassword.value = '';
  confirmPassword.value = '';
  errorMessage.value = '';
  successMessage.value = '';
};

const validatePassword = (value: string) => value.length >= 8 && /[A-Za-z]/.test(value) && /\d/.test(value);

watch(() => props.isOpen, (open) => {
  if (open) resetState();
});

const handleSubmit = async () => {
  errorMessage.value = '';
  successMessage.value = '';

  if (!validatePassword(newPassword.value.trim())) {
    errorMessage.value = PASSWORD_RULE;
    return;
  }
  if (newPassword.value.trim() !== confirmPassword.value.trim()) {
    errorMessage.value = '两次输入的新密码不一致';
    return;
  }
  if (oldPassword.value.trim() === newPassword.value.trim()) {
    errorMessage.value = '新密码不能与旧密码相同';
    return;
  }

  isLoading.value = true;
  try {
    await api.changePassword({
      oldPassword: oldPassword.value.trim(),
      newPassword: newPassword.value.trim(),
    });
    successMessage.value = '密码修改成功';
    emit('success');
    setTimeout(() => emit('close'), 600);
  } catch (error: any) {
    errorMessage.value = error.message || '密码修改失败';
  } finally {
    isLoading.value = false;
  }
};

const handleOpenChange = (open: boolean) => {
  if (!open) emit('close');
};
</script>

<template>
  <Dialog :open="isOpen" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[400px] p-0 overflow-hidden rounded-[20px] border border-white/60 shadow-[0_32px_64px_-12px_rgba(0,0,0,0.15)] bg-white/95 backdrop-blur-3xl ring-0">
      <div class="p-6">
        <DialogHeader class="mb-5 text-center sm:text-center">
          <DialogTitle class="text-xl font-black text-slate-800 tracking-tight leading-tight">
            修改密码
          </DialogTitle>
          <DialogDescription class="text-xs text-slate-500 font-medium mt-1.5 leading-relaxed">
            输入旧密码并设置一个新的安全密码。
          </DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">旧密码</label>
            <input
                v-model="oldPassword"
                type="password"
                :disabled="isLoading"
                autocomplete="current-password"
                class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
            />
          </div>

          <div class="space-y-1.5">
            <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">新密码</label>
            <input
                v-model="newPassword"
                type="password"
                :disabled="isLoading"
                autocomplete="new-password"
                class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
            />
            <p class="text-[11px] text-slate-400 ml-1">{{ PASSWORD_RULE }}</p>
          </div>

          <div class="space-y-1.5">
            <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">确认新密码</label>
            <input
                v-model="confirmPassword"
                type="password"
                :disabled="isLoading"
                autocomplete="new-password"
                class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
            />
          </div>

          <div v-if="errorMessage" class="rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-[12px] font-medium text-red-600">
            {{ errorMessage }}
          </div>
          <div v-if="successMessage" class="rounded-lg border border-emerald-100 bg-emerald-50 px-3 py-2 text-[12px] font-medium text-emerald-600">
            {{ successMessage }}
          </div>

          <button
              type="submit"
              class="w-full h-10 rounded-lg bg-blue-600 text-white font-bold text-sm shadow-md shadow-blue-200/50 hover:bg-blue-700 transition-all active:scale-[0.98] disabled:opacity-50 flex items-center justify-center gap-2"
              :disabled="!canSubmit"
          >
            <Loader2 v-if="isLoading" class="w-4 h-4 animate-spin" />
            <span>{{ isLoading ? '处理中...' : '确认修改' }}</span>
          </button>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>
