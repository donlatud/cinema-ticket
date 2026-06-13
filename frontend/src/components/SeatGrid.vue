<script setup>
import { computed } from 'vue'

const props = defineProps({
  seats: {
    type: Array,
    required: true,
  },
  selectedSeatNos: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['toggle'])

const columnNumbers = computed(() => {
  if (props.seats.length === 0) {
    return []
  }
  const maxCol = Math.max(
    ...props.seats.map((seat) => parseInt(seat.seat_no.slice(1), 10)),
  )
  return Array.from({ length: maxCol }, (_, index) => index + 1)
})

const rows = computed(() => {
  const grouped = {}
  for (const seat of props.seats) {
    const row = seat.seat_no.charAt(0)
    if (!grouped[row]) {
      grouped[row] = []
    }
    grouped[row].push(seat)
  }

  return Object.keys(grouped)
    .sort()
    .map((row) => ({
      row,
      seats: grouped[row].sort((a, b) => {
        const numA = parseInt(a.seat_no.slice(1), 10)
        const numB = parseInt(b.seat_no.slice(1), 10)
        return numA - numB
      }),
    }))
})

function seatClass(seat) {
  if (props.selectedSeatNos.includes(seat.seat_no)) {
    return 'seat-selected'
  }
  if (seat.status === 'LOCKED') {
    return 'seat-locked'
  }
  if (seat.status === 'BOOKED') {
    return 'seat-booked'
  }
  return 'seat-available'
}

function seatLabel(seat) {
  if (seat.status === 'AVAILABLE') {
    return `Seat ${seat.seat_no}, available`
  }
  if (seat.status === 'LOCKED') {
    return `Seat ${seat.seat_no}, locked`
  }
  return `Seat ${seat.seat_no}, booked`
}
</script>

<template>
  <section class="seatmap-grid" aria-label="Seat map">
    <div class="seatmap-grid__screen-wrap">
      <div class="seatmap-grid__screen" aria-hidden="true" />
      <p class="seatmap-grid__screen-label">SCREEN</p>
    </div>

    <div class="seatmap-grid__table" role="grid" aria-label="Seat selection grid">
      <div class="seatmap-grid__columns" role="row">
        <span class="seatmap-grid__corner" aria-hidden="true" />
        <span
          v-for="col in columnNumbers"
          :key="`col-${col}`"
          class="seatmap-grid__col-label"
          role="columnheader"
        >
          {{ col }}
        </span>
      </div>

      <div
        v-for="group in rows"
        :key="group.row"
        class="seatmap-grid__row"
        role="row"
        :aria-label="`Row ${group.row}`"
      >
        <span class="seatmap-grid__row-label" role="rowheader">{{ group.row }}</span>
        <button
          v-for="seat in group.seats"
          :key="seat.id"
          type="button"
          class="seatmap-grid__seat"
          :class="seatClass(seat)"
          :aria-label="seatLabel(seat)"
          :disabled="seat.status !== 'AVAILABLE'"
          @click="emit('toggle', seat.seat_no)"
        >
          {{ seat.seat_no }}
        </button>
      </div>
    </div>
  </section>
</template>
