<script setup>
defineProps({
  hasError: {
    type: Boolean,
    default: false,
  },
})
</script>

<template>
  <section
    class="w-full max-w-[400px] rounded-xl border bg-white p-8 shadow-sm sm:p-10"
    :class="hasError ? 'border-red-300' : 'border-slate-200'"
    aria-labelledby="login-title"
  >
    <slot />
  </section>
</template>
