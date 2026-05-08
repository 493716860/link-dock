<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { api } from '../api';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog';
import { Loader2 } from 'lucide-vue-next';

const props = defineProps<{
  isOpen: boolean;
  initialIsRegister?: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'login', user: { id: string; username: string }): void;
}>();

type Mode = 'login' | 'register' | 'forgot';

const USERNAME_RULE = '3-20 位字母、数字或下划线';
const USERNAME_TAKEN_MESSAGE = '该用户名已被占用';
const PASSWORD_RULE = '至少 8 位，且同时包含字母和数字';
const PRESET_QUESTIONS = [
  '你最常用的网站是？',
  '你第一次上网常去的网站是？',
  '你最喜欢的浏览器是什么？',
];

const mode = ref<Mode>('login');
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const securityQuestion = ref(PRESET_QUESTIONS[0]);
const customQuestion = ref('');
const securityAnswer = ref('');
const resetPassword = ref('');
const resetConfirmPassword = ref('');
const recoveryQuestion = ref('');
const isLoading = ref(false);
const statusMessage = ref('');
const errorMessage = ref('');
const usernameStatus = ref<'idle' | 'checking' | 'available' | 'taken' | 'invalid'>('idle');
let usernameCheckTimer: ReturnType<typeof setTimeout> | null = null;

const normalizedUsername = computed(() => username.value.trim().toLowerCase());
const effectiveSecurityQuestion = computed(() => securityQuestion.value === '__custom__' ? customQuestion.value.trim() : securityQuestion.value);

const modeMeta = computed(() => {
  if (mode.value === 'register') {
    return {
      title: '注册',
      description: '创建一个更安全的账户，用户名唯一且支持安全问题找回密码。',
    };
  }
  if (mode.value === 'forgot') {
    return {
      title: '找回密码',
      description: '通过用户名和安全问题答案重置密码，无需邮箱。',
    };
  }
  return {
    title: '登录',
    description: '欢迎回来，请登录以访问你的专属导航中心。',
  };
});

const canSubmit = computed(() => {
  if (isLoading.value) return false;
  if (mode.value === 'login') {
    return normalizedUsername.value.length > 0 && password.value.trim().length > 0;
  }
  if (mode.value === 'register') {
    return normalizedUsername.value.length > 0
        && password.value.trim().length > 0
        && confirmPassword.value.trim().length > 0
        && effectiveSecurityQuestion.value.length > 0
        && securityAnswer.value.trim().length > 0
        && usernameStatus.value === 'available';
  }
  return normalizedUsername.value.length > 0
      && recoveryQuestion.value.length > 0
      && securityAnswer.value.trim().length > 0
      && resetPassword.value.trim().length > 0
      && resetConfirmPassword.value.trim().length > 0;
});

const clearMessages = () => {
  statusMessage.value = '';
  errorMessage.value = '';
};

const resetFields = () => {
  username.value = '';
  password.value = '';
  confirmPassword.value = '';
  securityQuestion.value = PRESET_QUESTIONS[0];
  customQuestion.value = '';
  securityAnswer.value = '';
  resetPassword.value = '';
  resetConfirmPassword.value = '';
  recoveryQuestion.value = '';
  usernameStatus.value = 'idle';
  clearMessages();
};

const validateUsername = (value: string) => /^[a-zA-Z0-9_]{3,20}$/.test(value);
const validatePassword = (value: string) => value.length >= 8 && /[A-Za-z]/.test(value) && /\d/.test(value);

const switchMode = (nextMode: Mode) => {
  mode.value = nextMode;
  resetFields();
};

const checkUsernameAvailability = async () => {
  if (mode.value !== 'register') return;
  const value = normalizedUsername.value;
  if (!value) {
    usernameStatus.value = 'idle';
    return;
  }
  if (!validateUsername(value)) {
    usernameStatus.value = 'invalid';
    return;
  }
  if (value === 'admin') {
    usernameStatus.value = 'taken';
    errorMessage.value = USERNAME_TAKEN_MESSAGE;
    return;
  }

  usernameStatus.value = 'checking';
  try {
    const result = await api.checkUsername(value);
    usernameStatus.value = result.available ? 'available' : 'taken';
    if (!result.available) {
      errorMessage.value = result.message || USERNAME_TAKEN_MESSAGE;
    } else if (errorMessage.value.includes('用户名')) {
      errorMessage.value = '';
    }
  } catch (error: any) {
    usernameStatus.value = 'idle';
    errorMessage.value = error.message || '用户名校验失败';
  }
};

watch(() => props.isOpen, (newVal) => {
  if (newVal) {
    mode.value = props.initialIsRegister ? 'register' : 'login';
    resetFields();
  }
});

watch(normalizedUsername, () => {
  clearMessages();
  if (mode.value !== 'register') return;
  if (usernameCheckTimer) clearTimeout(usernameCheckTimer);
  usernameCheckTimer = setTimeout(() => {
    checkUsernameAvailability();
  }, 300);
});

const handleLogin = async () => {
  if (!normalizedUsername.value || !password.value.trim()) {
    errorMessage.value = '请输入用户名和密码';
    return;
  }

  isLoading.value = true;
  clearMessages();
  try {
    const user = await api.login({
      username: normalizedUsername.value,
      password: password.value.trim(),
    });
    emit('login', user);
    switchMode('login');
  } catch (error: any) {
    errorMessage.value = error.message || '登录失败';
  } finally {
    isLoading.value = false;
  }
};

const handleRegister = async () => {
  clearMessages();

  if (!validateUsername(normalizedUsername.value)) {
    errorMessage.value = `用户名需为 ${USERNAME_RULE}`;
    return;
  }
  if (normalizedUsername.value === 'admin') {
    errorMessage.value = USERNAME_TAKEN_MESSAGE;
    return;
  }
  if (usernameStatus.value !== 'available') {
    errorMessage.value = '请先确认用户名可用';
    return;
  }
  if (!validatePassword(password.value.trim())) {
    errorMessage.value = PASSWORD_RULE;
    return;
  }
  if (password.value.trim() !== confirmPassword.value.trim()) {
    errorMessage.value = '两次输入的密码不一致';
    return;
  }
  if (!effectiveSecurityQuestion.value) {
    errorMessage.value = '请选择或填写安全问题';
    return;
  }
  if (securityAnswer.value.trim().length < 2) {
    errorMessage.value = '安全问题答案至少需要 2 个字符';
    return;
  }

  isLoading.value = true;
  try {
    const user = await api.register({
      username: normalizedUsername.value,
      password: password.value.trim(),
      securityQuestion: effectiveSecurityQuestion.value,
      securityAnswer: securityAnswer.value.trim(),
    });
    emit('login', user);
    switchMode('login');
  } catch (error: any) {
    errorMessage.value = error.message || '注册失败';
  } finally {
    isLoading.value = false;
  }
};

const loadRecoveryQuestion = async () => {
  clearMessages();
  if (!normalizedUsername.value) {
    errorMessage.value = '请输入用户名';
    return;
  }

  isLoading.value = true;
  try {
    const result = await api.getRecoveryQuestion(normalizedUsername.value);
    recoveryQuestion.value = result.securityQuestion || '';
    statusMessage.value = '已找到安全问题，请完成验证后重置密码';
  } catch (error: any) {
    errorMessage.value = error.message || '获取安全问题失败';
  } finally {
    isLoading.value = false;
  }
};

const handleResetPassword = async () => {
  clearMessages();
  if (!recoveryQuestion.value) {
    errorMessage.value = '请先查询安全问题';
    return;
  }
  if (!validatePassword(resetPassword.value.trim())) {
    errorMessage.value = PASSWORD_RULE;
    return;
  }
  if (resetPassword.value.trim() !== resetConfirmPassword.value.trim()) {
    errorMessage.value = '两次输入的新密码不一致';
    return;
  }

  isLoading.value = true;
  try {
    await api.resetPassword({
      username: normalizedUsername.value,
      securityAnswer: securityAnswer.value.trim(),
      newPassword: resetPassword.value.trim(),
    });
    const resetSuccessMessage = '密码已重置，请使用新密码登录';
    password.value = '';
    switchMode('login');
    username.value = normalizedUsername.value;
    statusMessage.value = resetSuccessMessage;
  } catch (error: any) {
    errorMessage.value = error.message || '密码重置失败';
  } finally {
    isLoading.value = false;
  }
};

const handleSubmit = async () => {
  if (mode.value === 'register') {
    await handleRegister();
    return;
  }
  if (mode.value === 'forgot') {
    await handleResetPassword();
    return;
  }
  await handleLogin();
};

const handleOpenChange = (open: boolean) => {
  if (!open) {
    emit('close');
    switchMode(props.initialIsRegister ? 'register' : 'login');
  }
};
</script>

<template>
  <Dialog :open="isOpen" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[420px] p-0 overflow-hidden rounded-[20px] border border-white/60 shadow-[0_32px_64px_-12px_rgba(0,0,0,0.15)] bg-white/95 backdrop-blur-3xl ring-0">
      <div class="relative p-6">
        <DialogHeader class="mb-5 text-center sm:text-center">
          <DialogTitle class="text-xl font-black text-slate-800 tracking-tight leading-tight">
            {{ modeMeta.title }}
          </DialogTitle>
          <DialogDescription class="text-xs text-slate-500 font-medium mt-1.5 leading-relaxed">
            {{ modeMeta.description }}
          </DialogDescription>
        </DialogHeader>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div class="space-y-1.5">
            <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">用户名</label>
            <input
                id="username"
                v-model="username"
                placeholder="请输入用户名"
                :disabled="isLoading"
                autocomplete="username"
                class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
            />
            <p v-if="mode === 'register'" class="text-[11px] ml-1"
               :class="usernameStatus === 'available' ? 'text-emerald-600' : usernameStatus === 'taken' || usernameStatus === 'invalid' ? 'text-red-500' : 'text-slate-400'">
              <template v-if="usernameStatus === 'checking'">正在检查用户名可用性...</template>
              <template v-else-if="usernameStatus === 'available'">用户名可使用</template>
              <template v-else-if="usernameStatus === 'taken'">{{ USERNAME_TAKEN_MESSAGE }}</template>
              <template v-else>{{ USERNAME_RULE }}</template>
            </p>
          </div>

          <template v-if="mode === 'login'">
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">密码</label>
              <input
                  id="password"
                  type="password"
                  v-model="password"
                  placeholder="••••••••"
                  :disabled="isLoading"
                  autocomplete="current-password"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              />
            </div>
          </template>

          <template v-if="mode === 'register'">
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">密码</label>
              <input
                  type="password"
                  v-model="password"
                  placeholder="请设置密码"
                  :disabled="isLoading"
                  autocomplete="new-password"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              />
              <p class="text-[11px] text-slate-400 ml-1">{{ PASSWORD_RULE }}</p>
            </div>
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">确认密码</label>
              <input
                  type="password"
                  v-model="confirmPassword"
                  placeholder="请再次输入密码"
                  :disabled="isLoading"
                  autocomplete="new-password"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">安全问题</label>
              <select
                  v-model="securityQuestion"
                  :disabled="isLoading"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              >
                <option v-for="question in PRESET_QUESTIONS" :key="question" :value="question">{{ question }}</option>
                <option value="__custom__">自定义问题</option>
              </select>
              <input
                  v-if="securityQuestion === '__custom__'"
                  v-model="customQuestion"
                  placeholder="请输入你的安全问题"
                  :disabled="isLoading"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">安全问题答案</label>
              <input
                  v-model="securityAnswer"
                  placeholder="请牢记这个答案"
                  :disabled="isLoading"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none"
              />
            </div>
          </template>

          <template v-if="mode === 'forgot'">
            <button
                type="button"
                class="w-full h-10 rounded-lg bg-slate-100 text-slate-700 font-bold text-sm hover:bg-slate-200 transition-all active:scale-[0.98] disabled:opacity-50"
                :disabled="isLoading || !normalizedUsername"
                @click="loadRecoveryQuestion"
            >
              {{ recoveryQuestion ? '重新获取安全问题' : '获取安全问题' }}
            </button>
            <div v-if="recoveryQuestion" class="rounded-lg border border-blue-100 bg-blue-50/70 px-3 py-3">
              <p class="text-[11px] font-bold text-blue-500 uppercase tracking-widest mb-1">安全问题</p>
              <p class="text-sm font-semibold text-slate-700">{{ recoveryQuestion }}</p>
            </div>
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">答案</label>
              <input
                  v-model="securityAnswer"
                  placeholder="请输入安全问题答案"
                  :disabled="isLoading || !recoveryQuestion"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none disabled:opacity-60"
              />
            </div>
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">新密码</label>
              <input
                  type="password"
                  v-model="resetPassword"
                  placeholder="请输入新密码"
                  :disabled="isLoading || !recoveryQuestion"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none disabled:opacity-60"
              />
              <p class="text-[11px] text-slate-400 ml-1">{{ PASSWORD_RULE }}</p>
            </div>
            <div class="space-y-1.5">
              <label class="text-[11px] font-bold text-slate-500 uppercase tracking-widest ml-1">确认新密码</label>
              <input
                  type="password"
                  v-model="resetConfirmPassword"
                  placeholder="请再次输入新密码"
                  :disabled="isLoading || !recoveryQuestion"
                  class="w-full h-10 px-3 rounded-lg bg-slate-100/50 border border-slate-200 text-sm font-medium text-slate-800 placeholder:text-slate-400 focus:bg-white focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10 transition-all outline-none disabled:opacity-60"
              />
            </div>
          </template>

          <div v-if="errorMessage" class="rounded-lg border border-red-100 bg-red-50 px-3 py-2 text-[12px] font-medium text-red-600">
            {{ errorMessage }}
          </div>
          <div v-if="statusMessage" class="rounded-lg border border-emerald-100 bg-emerald-50 px-3 py-2 text-[12px] font-medium text-emerald-600">
            {{ statusMessage }}
          </div>

          <button
              type="submit"
              class="w-full h-10 mt-2 rounded-lg bg-blue-600 text-white font-bold text-sm shadow-md shadow-blue-200/50 hover:bg-blue-700 hover:-translate-y-px transition-all active:scale-[0.98] disabled:opacity-50 disabled:hover:translate-y-0 flex items-center justify-center gap-2"
              :disabled="!canSubmit"
          >
            <Loader2 v-if="isLoading" class="w-4 h-4 animate-spin" />
            <span>
              {{ isLoading ? '处理中...' : mode === 'register' ? '立即注册' : mode === 'forgot' ? '重置密码' : '登 录' }}
            </span>
          </button>

          <div class="pt-5 border-t border-slate-100/80 flex flex-col items-center gap-2.5 mt-5">
            <div v-if="mode === 'login'" class="flex items-center gap-3 text-[11px] font-bold">
              <button type="button" class="text-slate-400 hover:text-blue-600 transition-colors" @click="switchMode('forgot')">
                忘记密码
              </button>
              <span class="text-slate-200">|</span>
              <button type="button" class="text-blue-600 hover:text-blue-700 transition-colors" @click="switchMode('register')">
                创建新账户
              </button>
            </div>

            <div v-else-if="mode === 'register'" class="flex items-center gap-3 text-[11px] font-bold">
              <span class="text-slate-400">已经有账户了？</span>
              <button type="button" class="text-blue-600 hover:text-blue-700 transition-colors" @click="switchMode('login')">
                返回登录
              </button>
            </div>

            <div v-else class="flex items-center gap-3 text-[11px] font-bold">
              <button type="button" class="text-blue-600 hover:text-blue-700 transition-colors" @click="switchMode('login')">
                返回登录
              </button>
              <span class="text-slate-200">|</span>
              <button type="button" class="text-slate-400 hover:text-blue-600 transition-colors" @click="switchMode('register')">
                去注册
              </button>
            </div>
          </div>
        </form>
      </div>
    </DialogContent>
  </Dialog>
</template>
