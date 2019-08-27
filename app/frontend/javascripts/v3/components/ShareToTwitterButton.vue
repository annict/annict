<template>
  <span class="c-share-button-twitter">
    <span class="btn btn-sm u-btn-twitter" @click="openWindow">
      <div class="small">
        <i class="fab fa-twitter mr-1"></i>
        {{ $root.$t('noun.tweet') }}
      </div>
    </span>
  </span>
</template>

<script lang="ts">
  import $ from 'jquery';

  export default {
    props: {
      text: {
        type: String,
        required: true
      },
      url: {
        type: String,
        required: true
      },
      hashtags: {
        type: String
      }
    },

    setup(props, _context) {
      const tweetUrl = () => {
        const baseTweetUrl = 'https://twitter.com/intent/tweet'
        const params = $.param({
          text: `${props.text} | Annict`,
          url: props.url,
          hashtags: props.hashtags
        });

        return `${baseTweetUrl}?${params}`;
      }

      const openWindow = () => {
        const left = (screen.width - 640) / 2;
        const top = (screen.height - 480) / 2;

        window.open(
          tweetUrl(),
          '',
          `width=640,height=480,left=${left},top=${top}`
        );
      }

      return {
        openWindow
      }
    },
  };
</script>
