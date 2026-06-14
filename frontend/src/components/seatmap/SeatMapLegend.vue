<template>
  <section
    id="seat-legend"
    class="seatmap-legend"
    aria-label="Seat status legend"
  >
    <ul class="seatmap-legend__list">
      <li class="seatmap-legend__item">
        <span class="seatmap-legend__swatch seat-available" aria-hidden="true" />
        <span>Available</span>
      </li>
      <li class="seatmap-legend__item">
        <span class="seatmap-legend__swatch seat-selected" aria-hidden="true" />
        <span>Selected</span>
      </li>
      <li class="seatmap-legend__item">
        <span class="seatmap-legend__swatch seat-locked" aria-hidden="true" />
        <span>Locked</span>
      </li>
      <li class="seatmap-legend__item">
        <span class="seatmap-legend__swatch seat-booked" aria-hidden="true" />
        <span>Booked</span>
      </li>
    </ul>
  </section>
</template>
