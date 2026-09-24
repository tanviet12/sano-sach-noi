<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

type Variant = 'default' | 'secondary' | 'outline' | 'destructive'

const props = withDefaults(defineProps<{ variant?: Variant; class?: string }>(), {
  variant: 'default',
})

const variants: Record<Variant, string> = {
  default: 'border-transparent bg-primary text-primary-foreground',
  secondary: 'border-transparent bg-secondary text-secondary-foreground',
  outline: 'text-foreground border-border',
  destructive: 'border-transparent bg-destructive text-destructive-foreground',
}

const classes = computed(() =>
  cn(
    'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-medium',
    variants[props.variant],
    props.class,
  ),
)
</script>

<template>
  <span :class="classes"><slot /></span>
</template>
