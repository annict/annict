<template>
  <div class="c-empty p-4 text-center">
    <i class="far fa-meh"></i>
    <div class="h2 mt-3">
      {{ text }}
      <slot></slot>
    </div>
  </div>
</template>

<script lang="ts">
  export default {
    props: {
      text: {
        type: String,
        required: true
      },
    },
  }
</script>
