<template>
  <div :class="{ 'c-like-button d-inline-block u-fake-link': true, 'is-liked': isLiked }" @click="toggleLike">
    <i :class="{ 'far fa-heart': !isLiked, 'fas fa-heart': isLiked }"></i>
    <span class="count">
      {{ likesCount }}
    </span>
  </div>
</template>

<script lang="ts">
  import $ from 'jquery';
  import { value } from 'vue-function-api'

  import { LikeWorkRecordMutation, UnlikeWorkRecordMutation } from '../mutations'

  export default {
    props: {
      resourceName: {
        type: String,
        required: true,
      },
      resourceId: {
        type: String,
        required: true,
      },
      initLikesCount: {
        type: Number,
        required: true,
      },
      initIsLiked: {
        type: Boolean,
        required: true,
      },
      isSignedIn: {
        type: Boolean,
        default: false,
      },
    },

    setup(props, _context) {
      const likesCount = value(props.initLikesCount)
      const isLiked = value(props.initIsLiked)
      const isLoading = value(false)

      let likeMutation = null
      let unlikeMutation = null
      if (props.resourceName === 'WorkRecord') {
        likeMutation = new LikeWorkRecordMutation({ workRecordId: props.resourceId })
        unlikeMutation = new UnlikeWorkRecordMutation({ workRecordId: props.resourceId })
      }

      const toggleLike = async () => {
        if (!props.isSignedIn) {
          $('.c-sign-up-modal').modal('show')
          return
        }

        if (isLoading.value) {
          return
        }

        isLoading.value = true

        if (isLiked.value) {
          await unlikeMutation.execute()
          isLoading.value = false
          likesCount.value += -1
          isLiked.value = false
        } else {
          await likeMutation.execute()
          isLoading.value = false
          likesCount.value += 1
          isLiked.value = true
        }
      }

      return {
        toggleLike,
        likesCount,
        isLiked,
        isLoading
      }
    },
  };
</script>
