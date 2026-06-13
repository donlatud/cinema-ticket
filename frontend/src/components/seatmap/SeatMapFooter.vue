<script setup>
defineProps({
  selectedCount: {
    type: Number,
    default: 0,
  },
  selectedLabel: {
    type: String,
    default: '',
  },
  totalPrice: {
    type: Number,
    default: 0,
  },
  locking: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
})

defineEmits(['confirm'])
</script>

<template>
  <footer class="seatmap-footer">
    <section class="seatmap-footer__summary" aria-label="Selected seats summary">
      <p class="seatmap-footer__label">Selected Seats ({{ selectedCount }})</p>
      <p class="seatmap-footer__seats">{{ selectedLabel || 'None' }}</p>
    </section>

    <section class="seatmap-footer__price" aria-label="Total price">
      <p class="seatmap-footer__label">Total Price</p>
      <p class="seatmap-footer__amount">${{ totalPrice.toFixed(2) }}</p>
    </section>

    <button
      type="button"
      class="seatmap-footer__confirm"
      :disabled="disabled || locking"
      @click="$emit('confirm')"
    >
      {{ locking ? 'Confirming...' : 'Confirm Seats →' }}
    </button>
  </footer>
</template>
