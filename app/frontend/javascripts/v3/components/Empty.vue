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
  import { createComponent } from '@vue/composition-api'

  export default createComponent({
    props: {
      text: {
        type: String,
        required: true
      },
    },
  })
</script>
