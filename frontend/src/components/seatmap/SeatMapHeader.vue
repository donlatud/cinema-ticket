<script setup>
defineProps({
  movieTitle: {
    type: String,
    default: 'Showtime',
  },
  startTime: {
    type: String,
    default: '',
  },
  screen: {
    type: String,
    default: '',
  },
  pricePerSeat: {
    type: Number,
    default: 0,
  },
})
</script>

<template>
  <header class="seatmap-header">
    <div class="seatmap-header__main">
      <router-link to="/" class="seatmap-header__back" aria-label="Back to showtimes">
        <span class="seatmap-header__back-icon" aria-hidden="true">←</span>
        <span>{{ movieTitle }}</span>
      </router-link>
      <p class="seatmap-header__meta">
        <span>{{ startTime }}</span>
        <span class="seatmap-header__dot" aria-hidden="true">•</span>
        <span>{{ screen }}</span>
        <span class="seatmap-header__dot" aria-hidden="true">•</span>
        <span>${{ pricePerSeat.toFixed(2) }}/seat</span>
      </p>
    </div>
    <a class="seatmap-header__help" href="#seat-legend">Help</a>
  </header>
</template>
