<template>
  <div
    class="flex min-h-[48px] w-full items-center justify-center gap-3 rounded-lg border border-slate-200 bg-white px-4 py-3 sm:min-h-[52px]"
    aria-live="polite"
    aria-busy="true"
  >
    <span
      class="h-5 w-5 animate-spin rounded-full border-2 border-slate-200 border-t-slate-600"
      aria-hidden="true"
    />
    <span class="text-sm font-medium text-slate-600 sm:text-base">Signing in...</span>
  </div>
</template>
