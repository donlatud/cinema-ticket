<script setup>
import GoogleSignInButton from '@/components/auth/login/GoogleSignInButton.vue'
import SignInLoadingIndicator from '@/components/auth/login/SignInLoadingIndicator.vue'

defineProps({
  loading: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['sign-in'])
</script>

<template>
  <article class="mt-8">
    <GoogleSignInButton v-if="!loading" @click="$emit('sign-in')" />
    <SignInLoadingIndicator v-else />
  </article>
</template>
