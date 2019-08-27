<template>
  <span :class="'badge u-badge-' + state">
    <i :class="'mr-1 far fa-' + icons[state]"></i>
    {{ $root.$t(`models.record.ratingState.${state}`) }}
  </span>
</template>

<script lang="ts">
  import { value } from 'vue-function-api'

  export default {
    props: {
      initState: {
        type: String,
        required: true
      },
    },

    setup(props, _context) {
      const state = value(props.initState.toLowerCase())
      const icons = {
        great: 'heart',
        good: 'thumbs-up',
        average: 'meh',
        bad: 'thumbs-down'
      }

      return {
        state,
        icons
      }
    }
  }
</script>
