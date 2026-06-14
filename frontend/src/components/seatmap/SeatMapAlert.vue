<script setup>
defineProps({
  message: {
    type: String,
    required: true,
  },
})
</script>

<template>
  <section class="seatmap-alert" role="alert" aria-live="assertive">
    <h2 class="seatmap-alert__title">Booking Conflict Detected</h2>
    <p class="seatmap-alert__message">{{ message }}</p>
  </section>
</template>
