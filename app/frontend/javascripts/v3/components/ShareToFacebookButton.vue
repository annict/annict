<template>
  <span class="c-share-button-facebook">
    <span class="btn btn-sm u-btn-facebook" @click="openWindow">
      <div class="small">
        <i class="fab fa-facebook mr-1"></i>
        {{ $root.$t('noun.share') }}
      </div>
    </span>
  </span>
</template>

<script lang="ts">
  import $ from 'jquery';
  import { createComponent } from '@vue/composition-api'

  export default createComponent({
    props: {
      url: {
        type: String,
        required: true
      },
    },

    setup(props, _context) {
      const shareUrl = () => {
        const baseShareUrl = 'https://www.facebook.com/sharer/sharer.php'
        const params = $.param({
          u: props.url,
          display: 'popup',
          ref: 'plugin',
          src: 'like',
          kid_directed_site: 0,
          app_id: window.annConfig.facebook.appId,
        });

        return `${baseShareUrl}?${params}`;
      }

      const openWindow = () => {
        const left = (screen.width - 640) / 2;
        const top = (screen.height - 480) / 2;

        window.open(
          shareUrl(),
          '',
          `width=640,height=480,left=${left},top=${top}`
        );
      }

      return {
        openWindow
      }
    },
  })
</script>
