<template>
  <span :class="'badge u-badge-' + state.kind">
    <i :class="'mr-1 far fa-' + icons[state.kind]"></i>
    {{ $root.$t(`models.record.ratingState.${state.kind}`) }}
  </span>
</template>

<script lang="ts">
  import { createComponent, reactive } from '@vue/composition-api'

  export default createComponent({
    props: {
      initKind: {
        type: String,
        required: true
      },
    },

    setup(props, _context) {
      const state = reactive({
        kind: props.initKind.toLowerCase(),
      })
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
  })
</script>
