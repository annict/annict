<template>
  <transition name="app" mode="out-in">
    <div class="c-slot-list py-3">
      <div v-if="!state.isLoading" key="content">
        <div class="text-right">
          <select class="d-inline-block form-control form-control-sm">
            <option>放送日が新しい順</option>
          </select>
        </div>
        <template v-for="slot in state.slots">
          <div class="row">
            <div class="col pt-3">
              <div class="small mb-2">
                <span class="badge badge-info mr-2">
                  {{ slot.channel.name }}
                </span>
                <span class="u-text-green mr-2">
                  {{ slot.startedAt | formatDateTime }}
                </span>
                <span class="badge badge-secondary" v-if="slot.rebroadcast">
                  {{ $root.$t('noun.rebroadcast') }}
                </span>
              </div>
              <div class="d-flex mb-2">
                <div class="flex-shrink-1">
                  <a :href="'/track/works/' + slot.work.annictId">
                    <img :alt="slot.work.title" class="img-fluid img-thumbnail rounded" :src="slot.work.image.internalUrl" width="48">
                  </a>
                </div>
                <div class="w-100 ml-2">
                  <div>
                    <a href="/works/6254">
                      {{ slot.work.localTitle }}
                    </a>
                  </div>
                  <div>
                    <a :href="'/works/' + slot.work.annictId + '/episodes/' + slot.episode.annictId">
                      {{ slot.episode.numberText }}
                      <span class="ml-1" v-if="slot.episode.title">
                        {{ slot.episode.title }}
                      </span>
                    </a>
                  </div>
                </div>
              </div>
              <form class="row">
                <div class="col">
                  <div class="mb-2">
                    <textarea placeholder="" rows="1" name="episode_record[body]" class="form-control"></textarea>
                  </div>
                  <div class="row">
                    <div class="col"></div>
                    <div class="col">
                      <div class="text-right">
                        <div class="small text-mute"></div>
                      </div>
                    </div>
                  </div>
                </div>
                <div class="col u-flex-grow-0 pl-0">
                  <button type="button" class="btn btn-primary">
                    <i class="fas fa-edit mr-0" aria-hidden="true"></i>
                  </button>
                </div>
              </form>
            </div>
          </div>
        </template>
        <template v-if="state.slots.length === 0">
          ＼(^o^)／
        </template>
      </div>
      <div v-else key="loading">
        <div class="d-flex justify-content-center">
          <div class="c-loading">
            <div class="c-loading__core">
              Loading...
            </div>
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
  import {createComponent, onMounted, reactive} from '@vue/composition-api'

  import { FetchUserSlotsQuery } from '../queries';

  export default createComponent({
    setup(_props, _context) {
      const state = reactive({
        isLoading: true,
        slots: [],
      })

      onMounted(async () => {
        state.isLoading = true
        state.slots = await new FetchUserSlotsQuery().execute()
        console.log('state.slots: ', state.slots)
        state.isLoading = false
      })

      return {
        state,
      }
    },
  })
</script>
