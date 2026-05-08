<script setup lang="ts">
import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Search, LogOut, Menu, KeyRound, Database, Wrench } from 'lucide-vue-next';
import { computed } from 'vue';

const props = defineProps<{
  searchQuery: string;
  user: { id: string; username: string; isSuperAdmin?: boolean } | null;
}>();

const emit = defineEmits<{
  (e: 'update:searchQuery', value: string): void;
  (e: 'login'): void;
  (e: 'logout'): void;
  (e: 'toggle-sidebar'): void;
  (e: 'add-site'): void;
  (e: 'open-toolbox'): void;
  (e: 'change-password'): void;
  (e: 'manage-seed'): void;
}>();

const localSearchQuery = computed({
  get: () => props.searchQuery,
  set: (val) => emit('update:searchQuery', val)
});
</script>

<template>
  <header class="sticky top-0 z-30 grid h-16 grid-cols-[auto_1fr_auto] items-center border-b border-slate-200 bg-white px-6 md:grid-cols-[1fr_minmax(320px,448px)_1fr] md:px-10">
    <div class="flex items-center justify-start">
      <Button
          variant="ghost"
          size="icon"
          class="md:hidden mr-4 text-slate-500 hover:bg-slate-50 rounded-xl"
          @click="emit('toggle-sidebar')"
      >
        <Menu class="h-5 w-5" />
      </Button>
    </div>

    <div class="flex min-w-0 items-center justify-center">
      <div class="relative w-full flex items-center group">
        <Search class="absolute left-3.5 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-400 z-10 transition-colors group-focus-within:text-blue-500" />
        <input
            v-model="localSearchQuery"
            placeholder="搜索你的灵感库..."
            class="w-full pl-10 pr-12 h-10 bg-slate-50 border border-transparent text-sm font-medium focus:outline-none focus:ring-4 focus:ring-blue-500/10 focus:bg-white focus:border-blue-200 transition-all rounded-full shadow-sm placeholder:text-slate-400 text-slate-800"
        />
        <div v-if="!searchQuery" class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center gap-1 px-1.5 py-0.5 bg-slate-100 border border-slate-200 rounded-md text-[9px] font-bold text-slate-400 pointer-events-none hidden sm:flex">
          <span>⌘</span>
          <span>K</span>
        </div>
      </div>
    </div>

    <div class="flex items-center justify-end gap-2 ml-6">
      <button
          type="button"
          @click="emit('open-toolbox')"
          class="inline-flex h-10 items-center gap-2 rounded-full border border-slate-200 bg-white px-3 text-xs font-bold text-slate-600 shadow-sm transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600"
      >
        <Wrench class="h-4 w-4" />
        <span class="hidden lg:inline">工具箱</span>
      </button>
      <div v-if="user" class="flex items-center gap-3">

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <button class="h-10 px-3 rounded-full bg-white border border-slate-200 shadow-sm flex items-center gap-2 hover:bg-slate-50 transition-all ml-2">
              <div class="w-6 h-6 rounded-full bg-gradient-to-tr from-blue-500 to-cyan-500 flex items-center justify-center text-[10px] font-black text-white shadow-inner capitalize">
                {{ user.username?.charAt(0).toUpperCase() || 'U' }}
              </div>
              <span class="text-xs font-bold text-slate-700 hidden lg:inline-block pr-1">{{ user.username }}</span>
            </button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" class="w-48 rounded-2xl border-slate-200 bg-white p-2 shadow-xl">
            <DropdownMenuLabel class="flex flex-col py-2 px-3">
              <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-0.5">当前账户</span>
              <span class="text-sm font-black text-slate-800">{{ user.username }}</span>
            </DropdownMenuLabel>
            <DropdownMenuSeparator class="my-1 border-slate-100" />
            <DropdownMenuItem @click="emit('change-password')" class="rounded-xl px-3 py-2 text-xs font-bold text-slate-700 hover:bg-slate-50 cursor-pointer transition-colors">
              <KeyRound class="w-4 h-4 mr-2" /> 修改密码
            </DropdownMenuItem>
            <DropdownMenuItem
                v-if="user.isSuperAdmin"
                @click="emit('manage-seed')"
                class="rounded-xl px-3 py-2 text-xs font-bold text-blue-600 hover:bg-blue-50 cursor-pointer transition-colors"
            >
              <Database class="w-4 h-4 mr-2" /> 默认书签管理
            </DropdownMenuItem>
            <DropdownMenuItem @click="emit('logout')" class="rounded-xl px-3 py-2 text-xs font-bold text-red-600 hover:bg-red-50 cursor-pointer transition-colors">
              <LogOut class="w-4 h-4 mr-2" /> 退出登录
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div v-else class="flex items-center gap-2">
        <Button
            class="rounded-full font-bold h-10 px-6 text-xs shadow-md shadow-blue-200/50 bg-blue-600 text-white hover:bg-blue-700 hover:-translate-y-0.5 transition-all active:scale-95"
            @click="emit('login')"
        >
          登录
        </Button>
      </div>
    </div>
  </header>
</template>
