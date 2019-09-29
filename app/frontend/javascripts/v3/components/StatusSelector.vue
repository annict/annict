<template>
  <div :class="{ 'c-status-selector': true, 'is-small': size === 'small', 'is-transparent': isTransparent(), 'c-spinner': state.isLoading }">
    <div class="c-status-selector__object">
      <select class="w-100" v-model="state.kind" @change="changeStatus">
        <option value="NO_STATUS">
          {{ $root.$t('messages._components.statusSelector.selectStatus') }}
        </option>
        <option value="PLAN_TO_WATCH">
          {{ $root.$t('models.status.kind.planToWatch') }}
        </option>
        <option value="WATCHING">
          {{ $root.$t('models.status.kind.watching') }}
        </option>
        <option value="COMPLETED">
          {{ $root.$t('models.status.kind.completed') }}
        </option>
        <option value="ON_HOLD">
          {{ $root.$t('models.status.kind.onHold') }}
        </option>
        <option value="DROPPED">
          {{ $root.$t('models.status.kind.dropped') }}
        </option>
      </select>
      <i class="fas fa-caret-down"></i>
    </div>
  </div>
</template>

<script lang="ts">
import $ from 'jquery'
import { createComponent, onMounted, reactive } from '@vue/composition-api'

import { UpdateStatusMutation } from '../mutations'

export default createComponent({
  props: {
    workId: {
      type: String,
      required: true,
    },

    size: {
      type: String,
      default: 'default',
    },

    initIsTransparent: {
      type: Boolean,
      default: false,
    },

    initKind: {
      type: String,
    },
  },

  setup(props, context) {
    const NO_STATUS = 'NO_STATUS'
    const state = reactive({
      appData: {},
      pageData: {},
      statuses: [],
      isLoading: false,
      kind: props.initKind,
      prevKind: null,
    })

    const resetKind = () => {
      state.kind = NO_STATUS
    }

    const changeStatus = async () => {
      if (!context.root.isSignedIn()) {
        $('.c-sign-up-modal').modal('show')
        resetKind()
        return
      }

      if (state.kind !== state.prevKind) {
        state.isLoading = true
        await new UpdateStatusMutation({ workId: props.workId, kind: state.kind }).execute()
        state.isLoading = false
      }
    }

    const isTransparent = () => {
      return props.initIsTransparent || !context.root.isSignedIn() || state.kind === NO_STATUS
    }

    onMounted(async () => {
      state.isLoading = true

      if (!context.root.isSignedIn()) {
        state.kind = state.prevKind = NO_STATUS
        state.isLoading = false
        return
      }

      if (props.initKind) {
        state.prevKind = props.initKind
        state.isLoading = false
        return
      }
    })

    return {
      state,
      changeStatus,
      isTransparent,
    }
  },
})
</script>
