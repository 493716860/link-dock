/// <reference types="vite/client" />

declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<{}, {}, any>;
  export default component;
}

interface ImportMetaEnv {
    readonly VITE_USE_MOCK: string
    // 如果以后还有其他的环境变量，也可以加在这里
}

interface ImportMeta {
    readonly env: ImportMetaEnv
}