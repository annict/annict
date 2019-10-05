<template>
  <div :class="{ 'c-like-button d-inline-block u-fake-link': true, 'is-liked': state.isLiked }" @click="toggleLike">
    <i :class="{ 'far fa-heart': !state.isLiked, 'fas fa-heart': state.isLiked }"></i>
    <span class="count">
      {{ state.likesCount }}
    </span>
  </div>
</template>

<script lang="ts">
  import $ from 'jquery';
  import { createComponent, reactive } from '@vue/composition-api'

  import { LikeWorkRecordMutation, UnlikeWorkRecordMutation } from '../mutations'

  export default createComponent({
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
      const state = reactive({
        likesCount: props.initLikesCount,
        isLiked: props.initIsLiked,
        isLoading: false,
      })

      let likeMutation = null
      let unlikeMutation = null
      if (props.resourceName === 'WorkRecord') {
        likeMutation = new LikeWorkRecordMutation({ workRecordId: props.resourceId })
        unlikeMutation = new UnlikeWorkRecordMutation({ workRecordId: props.resourceId })
      }

      const toggleLike = async () => {
        if (!props.isSignedIn) {
          ($('.c-sign-up-modal') as any).modal('show')
          return
        }

        if (state.isLoading) {
          return
        }

        state.isLoading = true

        if (state.isLiked) {
          await unlikeMutation.execute()
          state.isLoading = false
          state.likesCount += -1
          state.isLiked = false
        } else {
          await likeMutation.execute()
          state.isLoading = false
          state.likesCount += 1
          state.isLiked = true
        }
      }

      return {
        toggleLike,
        state,
      }
    },
  })
</script>
