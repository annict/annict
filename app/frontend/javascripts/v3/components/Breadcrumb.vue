<template>
  <div class="c-breadcrumb">
    <template v-for="item in items">
      <template v-if="item.current">
        <span class="current">
          {{ item.text }}
        </span>
      </template>
      <template v-else>
        <a :href="item.href">
          {{ item.text }}
        </a>
        <span class="px-1">›</span>
      </template>
    </template>
  </div>
</template>

<script lang="ts">
  import { createComponent } from '@vue/composition-api'

  export default createComponent({
    props: {
      items: {
        type: Array,
        required: true
      },
    },
  })
</script>
