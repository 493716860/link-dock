<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue';
import { Check, Download, FileImage, ImagePlus, ListChecks, NotebookPen, Plus, Trash2 } from 'lucide-vue-next';

type ToolTab = 'notes' | 'todos' | 'images';

type NotebookNote = {
  id: string;
  title: string;
  content: string;
  createdAt: string;
  updatedAt: string;
};

type TodoItem = {
  id: string;
  title: string;
  done: boolean;
  dueText: string;
  createdAt: string;
};

type ConvertedImage = {
  name: string;
  size: number;
  url: string;
  type: string;
};

const props = defineProps<{
  currentPageData: { url?: string; name?: string; description?: string } | null;
}>();

const activeTab = ref<ToolTab>('notes');
const notes = ref<NotebookNote[]>([]);
const todos = ref<TodoItem[]>([]);
const activeNoteId = ref('');
const noteSearch = ref('');
const todoTitle = ref('');
const todoDueText = ref('');
const imageFormat = ref<'image/webp' | 'image/jpeg' | 'image/png'>('image/webp');
const imageQuality = ref(0.82);
const convertedImage = ref<ConvertedImage | null>(null);
const imageError = ref('');
const isConverting = ref(false);

const NOTES_KEY = 'linkdock-office-notes';
const TODOS_KEY = 'linkdock-office-todos';

const toolTabs = [
  { id: 'notes', label: '记事本', icon: NotebookPen },
  { id: 'todos', label: '备忘待办', icon: ListChecks },
  { id: 'images', label: '图片工具', icon: FileImage },
] as const;

const todayTitle = computed(() => {
  const formatter = new Intl.DateTimeFormat('zh-CN', { month: '2-digit', day: '2-digit', weekday: 'short' });
  return `${formatter.format(new Date())} 工作计划`;
});

const formatBytes = (bytes: number) => {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(2)} MB`;
};

const makeId = (prefix: string) => `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 7)}`;

const loadLocalData = () => {
  try {
    notes.value = JSON.parse(window.localStorage.getItem(NOTES_KEY) || '[]').map((note: any) => ({
      id: note.id,
      title: note.title || '未命名笔记',
      content: note.content || '',
      createdAt: note.createdAt || new Date().toISOString(),
      updatedAt: note.updatedAt || note.createdAt || new Date().toISOString(),
    }));
    todos.value = JSON.parse(window.localStorage.getItem(TODOS_KEY) || '[]');
  } catch {
    notes.value = [];
    todos.value = [];
  }
};

onMounted(() => {
  loadLocalData();
  if (notes.value.length === 0) {
    createNote(todayTitle.value, '今日计划：\n\n- ');
    return;
  }
  activeNoteId.value = notes.value[0]?.id || '';
});

watch(notes, value => {
  window.localStorage.setItem(NOTES_KEY, JSON.stringify(value));
}, { deep: true });

watch(todos, value => {
  window.localStorage.setItem(TODOS_KEY, JSON.stringify(value));
}, { deep: true });

const filteredNotes = computed(() => {
  const keyword = noteSearch.value.trim().toLowerCase();
  const sorted = [...notes.value].sort((a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime());
  if (!keyword) return sorted;
  return sorted.filter(note => {
    return [note.title, note.content].some(value => value.toLowerCase().includes(keyword));
  });
});

const activeNote = computed(() => notes.value.find(note => note.id === activeNoteId.value) || notes.value[0] || null);

const openTodos = computed(() => todos.value.filter(todo => !todo.done));
const doneTodos = computed(() => todos.value.filter(todo => todo.done));

const createNote = (title = '未命名笔记', content = '') => {
  const now = new Date().toISOString();
  const note = {
    id: makeId('note'),
    title,
    content,
    createdAt: now,
    updatedAt: now,
  };
  notes.value.unshift(note);
  activeNoteId.value = note.id;
};

const createTodayPlan = () => {
  const existing = notes.value.find(note => note.title === todayTitle.value);
  if (existing) {
    activeNoteId.value = existing.id;
    return;
  }
  createNote(todayTitle.value, '今日计划：\n\n- ');
};

const deleteNote = (id: string) => {
  notes.value = notes.value.filter(note => note.id !== id);
  if (activeNoteId.value === id) {
    activeNoteId.value = notes.value[0]?.id || '';
  }
};

const updateActiveNote = (field: 'title' | 'content', value: string) => {
  if (!activeNote.value) return;
  const now = new Date().toISOString();
  notes.value = notes.value.map(note => note.id === activeNote.value?.id ? {
    ...note,
    [field]: value,
    updatedAt: now,
  } : note);
};

const addTodo = () => {
  const title = todoTitle.value.trim();
  if (!title) return;
  todos.value.unshift({
    id: makeId('todo'),
    title,
    done: false,
    dueText: todoDueText.value.trim(),
    createdAt: new Date().toISOString(),
  });
  todoTitle.value = '';
  todoDueText.value = '';
};

const toggleTodo = (id: string) => {
  todos.value = todos.value.map(todo => todo.id === id ? { ...todo, done: !todo.done } : todo);
};

const deleteTodo = (id: string) => {
  todos.value = todos.value.filter(todo => todo.id !== id);
};

const convertImage = async (event: Event) => {
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  imageError.value = '';
  isConverting.value = true;
  if (convertedImage.value?.url) URL.revokeObjectURL(convertedImage.value.url);
  convertedImage.value = null;

  try {
    const bitmap = await createImageBitmap(file);
    const canvas = document.createElement('canvas');
    canvas.width = bitmap.width;
    canvas.height = bitmap.height;
    const context = canvas.getContext('2d');
    if (!context) throw new Error('无法创建图片画布');
    context.drawImage(bitmap, 0, 0);
    bitmap.close();

    const blob = await new Promise<Blob>((resolve, reject) => {
      canvas.toBlob(result => result ? resolve(result) : reject(new Error('图片转换失败')), imageFormat.value, imageQuality.value);
    });

    const extension = imageFormat.value.split('/')[1];
    const sourceName = file.name.replace(/\.[^.]+$/, '');
    convertedImage.value = {
      name: `${sourceName}.${extension}`,
      size: blob.size,
      url: URL.createObjectURL(blob),
      type: imageFormat.value,
    };
  } catch (error: any) {
    imageError.value = error?.message || '图片处理失败';
  } finally {
    isConverting.value = false;
    input.value = '';
  }
};
</script>

<template>
  <div class="mx-auto w-full max-w-6xl">
    <div class="mb-6 flex flex-col gap-4 border-b border-slate-200/70 pb-5 md:flex-row md:items-end md:justify-between">
      <div>
        <p class="text-[11px] font-black uppercase tracking-[0.24em] text-blue-600">Toolbox</p>
        <h1 class="mt-1 text-2xl font-black tracking-tight text-slate-900">办公工具箱</h1>
      </div>
      <div class="flex flex-wrap gap-2">
        <button
            v-for="tab in toolTabs"
            :key="tab.id"
            type="button"
            @click="activeTab = tab.id"
            class="inline-flex h-9 items-center gap-2 rounded-xl border px-3 text-xs font-bold transition-all"
            :class="activeTab === tab.id ? 'border-blue-200 bg-blue-50 text-blue-600' : 'border-slate-200 bg-white text-slate-500 hover:border-slate-300 hover:text-slate-800'"
        >
          <component :is="tab.icon" class="h-3.5 w-3.5" />
          {{ tab.label }}
        </button>
      </div>
    </div>

    <section v-if="activeTab === 'notes'" class="grid min-h-[620px] gap-5 lg:grid-cols-[320px_1fr]">
      <aside class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <div class="mb-3 flex gap-2">
          <button @click="createNote()" class="inline-flex h-9 flex-1 items-center justify-center gap-2 rounded-xl bg-blue-600 px-3 text-xs font-bold text-white transition-all hover:bg-blue-700">
            <Plus class="h-3.5 w-3.5" />
            新建
          </button>
          <button @click="createTodayPlan" class="inline-flex h-9 flex-1 items-center justify-center rounded-xl border border-slate-200 bg-white px-3 text-xs font-bold text-slate-600 transition-all hover:border-blue-100 hover:bg-blue-50 hover:text-blue-600">
            今日计划
          </button>
        </div>
        <input v-model="noteSearch" class="mb-3 h-9 w-full rounded-xl border border-slate-200 bg-slate-50 px-3 text-xs font-bold text-slate-600 outline-none focus:border-blue-300 focus:bg-white focus:ring-4 focus:ring-blue-500/10" placeholder="搜索记事本" />
        <div class="max-h-[520px] space-y-2 overflow-y-auto pr-1 custom-scrollbar">
          <div v-if="filteredNotes.length === 0" class="rounded-2xl bg-slate-50 px-4 py-10 text-center text-xs font-bold text-slate-400">
            暂时还没有笔记
          </div>
          <button
              v-for="note in filteredNotes"
              :key="note.id"
              type="button"
              @click="activeNoteId = note.id"
              class="block w-full rounded-2xl border p-3 text-left transition-all"
              :class="activeNote?.id === note.id ? 'border-blue-200 bg-blue-50/70 text-blue-700' : 'border-slate-100 bg-slate-50/70 text-slate-700 hover:border-slate-200 hover:bg-white'"
          >
            <p class="truncate text-sm font-black">{{ note.title || '未命名笔记' }}</p>
            <p class="mt-1 line-clamp-2 text-[11px] leading-relaxed opacity-70">{{ note.content || '空白笔记' }}</p>
          </button>
        </div>
      </aside>

      <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <div v-if="activeNote" class="flex h-full min-h-[560px] flex-col">
          <div class="mb-3 flex items-center justify-between gap-3 border-b border-slate-100 pb-3">
            <input
                :value="activeNote.title"
                @input="updateActiveNote('title', ($event.target as HTMLInputElement).value)"
                class="h-11 min-w-0 flex-1 rounded-xl border border-transparent bg-transparent px-1 text-xl font-black text-slate-900 outline-none focus:border-blue-200 focus:bg-slate-50 focus:px-3 focus:ring-4 focus:ring-blue-500/10"
                placeholder="未命名笔记"
            />
            <div class="flex shrink-0 items-center gap-2">
              <span class="hidden rounded-full bg-emerald-50 px-2.5 py-1 text-[10px] font-black text-emerald-600 sm:inline-flex">自动保存</span>
              <button @click="deleteNote(activeNote.id)" class="inline-flex h-9 w-9 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-red-50 hover:text-red-500">
                <Trash2 class="h-4 w-4" />
              </button>
            </div>
          </div>
          <textarea
              :value="activeNote.content"
              @input="updateActiveNote('content', ($event.target as HTMLTextAreaElement).value)"
              class="min-h-0 flex-1 resize-none rounded-2xl border border-slate-100 bg-slate-50/70 px-4 py-4 text-sm leading-7 text-slate-700 outline-none focus:border-blue-200 focus:bg-white focus:ring-4 focus:ring-blue-500/10"
              placeholder="写下今天的工作计划、会议记录、文案草稿或临时想法..."
          ></textarea>
        </div>
        <div v-else class="flex min-h-[560px] items-center justify-center rounded-2xl bg-slate-50 text-xs font-bold text-slate-400">
          新建一篇笔记后开始记录
        </div>
      </div>
    </section>

    <section v-else-if="activeTab === 'todos'" class="grid gap-5 lg:grid-cols-[minmax(320px,0.75fr)_1.25fr]">
      <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h2 class="mb-4 text-sm font-black text-slate-800">新增备忘</h2>
        <div class="space-y-3">
          <input v-model="todoTitle" @keyup.enter="addTodo" class="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm font-bold text-slate-800 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" placeholder="要处理什么事？" />
          <input v-model="todoDueText" @keyup.enter="addTodo" class="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm text-slate-600 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10" placeholder="时间备注，例如：今天 18:00" />
          <button @click="addTodo" class="inline-flex h-10 w-full items-center justify-center gap-2 rounded-xl bg-blue-600 text-xs font-bold text-white shadow-md shadow-blue-200/60 transition-all hover:bg-blue-700">
            <Plus class="h-4 w-4" />
            添加备忘
          </button>
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <div class="mb-4 flex items-center justify-between">
          <h2 class="text-sm font-black text-slate-800">待办列表</h2>
          <span class="rounded-full bg-slate-100 px-2.5 py-1 text-[10px] font-black text-slate-500">{{ openTodos.length }} OPEN</span>
        </div>
        <div class="space-y-2">
          <div v-if="todos.length === 0" class="rounded-2xl bg-slate-50 px-4 py-10 text-center text-xs font-bold text-slate-400">
            暂时没有备忘事项
          </div>
          <div v-for="todo in [...openTodos, ...doneTodos]" :key="todo.id" class="flex items-center gap-3 rounded-2xl border border-slate-100 bg-slate-50/70 p-3">
            <button @click="toggleTodo(todo.id)" class="flex h-8 w-8 shrink-0 items-center justify-center rounded-xl border transition-all" :class="todo.done ? 'border-emerald-200 bg-emerald-50 text-emerald-600' : 'border-slate-200 bg-white text-slate-300 hover:border-blue-200 hover:text-blue-600'">
              <Check class="h-4 w-4" />
            </button>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm font-bold" :class="todo.done ? 'text-slate-400 line-through' : 'text-slate-800'">{{ todo.title }}</p>
              <p v-if="todo.dueText" class="mt-0.5 truncate text-[11px] font-semibold text-slate-400">{{ todo.dueText }}</p>
            </div>
            <button @click="deleteTodo(todo.id)" class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-red-50 hover:text-red-500">
              <Trash2 class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>
    </section>

    <section v-else class="grid gap-5 lg:grid-cols-[minmax(320px,0.85fr)_1.15fr]">
      <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h2 class="mb-4 text-sm font-black text-slate-800">图片压缩与格式转换</h2>
        <div class="space-y-4">
          <label class="flex min-h-44 cursor-pointer flex-col items-center justify-center rounded-2xl border border-dashed border-slate-300 bg-slate-50 px-4 text-center transition-all hover:border-blue-300 hover:bg-blue-50/50">
            <ImagePlus class="mb-3 h-8 w-8 text-slate-300" />
            <span class="text-sm font-black text-slate-700">选择图片</span>
            <span class="mt-1 text-xs text-slate-400">本地处理，不上传文件</span>
            <input class="hidden" type="file" accept="image/*" @change="convertImage" />
          </label>
          <div class="grid gap-3 sm:grid-cols-2">
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">目标格式</span>
              <select v-model="imageFormat" class="h-10 w-full rounded-xl border border-slate-200 bg-white px-3 text-sm font-bold text-slate-700 outline-none focus:border-blue-300 focus:ring-4 focus:ring-blue-500/10">
                <option value="image/webp">WebP</option>
                <option value="image/jpeg">JPEG</option>
                <option value="image/png">PNG</option>
              </select>
            </label>
            <label class="space-y-1.5">
              <span class="text-[10px] font-black uppercase tracking-widest text-slate-400">质量 {{ Math.round(imageQuality * 100) }}%</span>
              <input v-model.number="imageQuality" type="range" min="0.3" max="1" step="0.02" class="h-10 w-full accent-blue-600" />
            </label>
          </div>
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h2 class="mb-4 text-sm font-black text-slate-800">处理结果</h2>
        <div v-if="isConverting" class="rounded-2xl bg-slate-50 px-4 py-12 text-center text-xs font-bold text-slate-400">正在处理图片...</div>
        <div v-else-if="imageError" class="rounded-2xl border border-red-100 bg-red-50 px-4 py-4 text-xs font-bold text-red-600">{{ imageError }}</div>
        <div v-else-if="convertedImage" class="rounded-2xl border border-slate-100 bg-slate-50/70 p-4">
          <p class="text-sm font-black text-slate-800">{{ convertedImage.name }}</p>
          <p class="mt-1 text-xs font-semibold text-slate-400">{{ convertedImage.type }} · {{ formatBytes(convertedImage.size) }}</p>
          <a :href="convertedImage.url" :download="convertedImage.name" class="mt-4 inline-flex h-10 items-center gap-2 rounded-xl bg-blue-600 px-4 text-xs font-bold text-white shadow-md shadow-blue-200/60 transition-all hover:bg-blue-700">
            <Download class="h-4 w-4" />
            下载图片
          </a>
        </div>
        <div v-else class="rounded-2xl bg-slate-50 px-4 py-12 text-center text-xs font-bold text-slate-400">处理完成后会在这里显示下载入口</div>
      </div>
    </section>
  </div>
</template>
