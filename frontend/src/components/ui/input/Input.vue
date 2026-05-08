<script setup lang="ts">
import type { HTMLAttributes } from "vue"
import { useVModel } from "@vueuse/core"
import { cn } from "@/lib/utils"

const props = defineProps<{
  defaultValue?: string | number
  modelValue?: string | number
  class?: HTMLAttributes["class"]
}>()

const emits = defineEmits<{
  (e: "update:modelValue", payload: string | number): void
}>()

const modelValue = useVModel(props, "modelValue", emits, {
  passive: true,
  defaultValue: props.defaultValue,
})
</script>

<template>
  <input
    v-model="modelValue"
    data-slot="input"
    :class="cn(
      'file:text-foreground placeholder:text-muted-foreground/30 selection:bg-black selection:text-white h-11 w-full min-w-0 rounded-none border border-border bg-transparent px-4 py-2 text-sm transition-all focus:outline-none focus:border-black disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 font-body',
      'aria-invalid:border-red-500',
      props.class,
    )"
  >
</template>
