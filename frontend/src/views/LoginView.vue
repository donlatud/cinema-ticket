<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginWithGoogle } from '@/firebase'
import { useAuthStore } from '@/stores/auth'
import LoginPageLayout from '@/components/auth/login/LoginPageLayout.vue'
import LoginCard from '@/components/auth/login/LoginCard.vue'
import LoginBrandHeader from '@/components/auth/login/LoginBrandHeader.vue'
import LoginErrorAlert from '@/components/auth/login/LoginErrorAlert.vue'
import LoginSignInAction from '@/components/auth/login/LoginSignInAction.vue'
import LoginFooter from '@/components/auth/login/LoginFooter.vue'

const router = useRouter()
const authStore = useAuthStore()

const status = ref('idle')
const errorMessage = ref('')

const isLoading = computed(() => status.value === 'loading')
const hasError = computed(() => status.value === 'error')

async function handleGoogleSignIn() {
  if (isLoading.value) {
    return
  }

  status.value = 'loading'
  errorMessage.value = ''

  try {
    const idToken = await loginWithGoogle()
    await authStore.loginWithIdToken(idToken)
    router.push('/')
  } catch {
    status.value = 'error'
    errorMessage.value = 'Sign-in failed. Please try again.'
  }
}
</script>

<template>
  <LoginPageLayout>
    <LoginCard :has-error="hasError">
      <LoginBrandHeader />

      <LoginErrorAlert v-if="hasError" :message="errorMessage" />

      <LoginSignInAction
        :loading="isLoading"
        @sign-in="handleGoogleSignIn"
      />

      <LoginFooter />
    </LoginCard>
  </LoginPageLayout>
</template>
