import type { VariantProps } from "class-variance-authority"
import { cva } from "class-variance-authority"

export { default as Button } from "./Button.vue"

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-3 whitespace-nowrap rounded-none text-[10px] font-bold uppercase tracking-[0.1em] transition-all disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-3.5 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:ring-1 focus-visible:ring-black",
  {
    variants: {
      variant: {
        default:
          "bg-black text-white hover:bg-neutral-800",
        destructive:
          "bg-red-500 text-white hover:bg-red-600",
        outline:
          "border border-border bg-transparent hover:bg-white hover:border-black",
        secondary:
          "bg-zinc-100 text-black hover:bg-zinc-200",
        ghost:
          "hover:bg-zinc-100 dark:hover:bg-zinc-800",
        link: "text-black underline-offset-4 hover:underline",
      },
      size: {
        "default": "h-11 px-8 py-2",
        "sm": "h-8 px-4",
        "lg": "h-14 px-12",
        "icon": "size-10",
        "icon-sm": "size-8",
        "icon-lg": "size-12",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)
export type ButtonVariants = VariantProps<typeof buttonVariants>
