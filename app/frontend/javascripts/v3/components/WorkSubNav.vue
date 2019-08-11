<template>
  <div class="c-subnav c-subnav--transparent">
    <a href="" :class="{ 'c-subnav__link': true, 'c-subnav__link--active': pageCategory === 'workDetail' }">
      <div class="c-subnav__item">
        {{ $t('noun.detail') }}
      </div>
    </a>
    <template v-if="!work.isNoEpisodes">
      <a href="" :class="{ 'c-subnav__link': true, 'c-subnav__link--active': pageCategory === 'episodeList' }">
        <div class="c-subnav__item">
          {{ $t('noun.episodes') }}
        </div>
      </a>
    </template>
    <a href="" :class="{ 'c-subnav__link': true, 'c-subnav__link--active': pageCategory === 'workRecordList' }">
      <div class="c-subnav__item">
        {{ $t('noun.records') }}
      </div>
    </a>
  </div>
</template>

<script lang="ts">
  export default {
    props: {
      work: {
        type: Object,
        required: true
      },
      pageCategory: {
        type: String,
        required: true
      }
    },
  }
</script>
