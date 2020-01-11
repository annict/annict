<template>
  <transition name="app" mode="out-in">
    <div v-if="state.libraryEntries" key="content">
      <ann-navbar></ann-navbar>
      <div class="container p-3" data-turbolinks="false">
        <ann-breadcrumb :items="state.breadcrumbItems" class="mb-3"></ann-breadcrumb>
        <div class="container">
          <div class="row">
            <div class="col-4">
              <div class="list-group mb-3">
                <a class="list-group-item list-group-item-action active" href="/track">
                  {{ $root.$t('noun.slots') }}
                </a>
              </div>
              <div class="list-group">
                <a class="align-items-center d-flex justify-content-between list-group-item list-group-item-action" :href="'/works/' + libraryEntry.work.annictId" v-for="libraryEntry in state.libraryEntries">
                  {{ libraryEntry.work.title }}
                  <span class="badge badge-primary badge-pill">
                    {{ libraryEntry.untappedEpisodesCount }}
                  </span>
                </a>
              </div>
            </div>
            <div class="col-8">
              <div class="c-card px-3">
                <ann-slot-list></ann-slot-list>
              </div>
            </div>
          </div>
        </div>
      </div>
      <ann-footer></ann-footer>
    </div>
    <div v-else key="loading">
      <div class="d-flex justify-content-center align-items-center vh-100">
        <div class="c-loading">
          <div class="c-loading__core">
            Loading...
          </div>
        </div>
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
  import { createComponent, onMounted, reactive } from '@vue/composition-api'

  import Breadcrumb from '../components/Breadcrumb.vue'
  import Empty from '../components/Empty.vue'
  import Footer from '../components/Footer.vue'
  import NavBar from '../components/NavBar.vue'
  import SlotList from '../components/SlotList.vue'
  import StatusSelector from '../components/StatusSelector.vue'

  import { FetchLibraryEntriesQuery } from '../queries'

  export default createComponent({
    components: {
      'ann-breadcrumb': Breadcrumb,
      'ann-empty': Empty,
      'ann-footer': Footer,
      'ann-navbar': NavBar,
      'ann-slot-list': SlotList,
      'ann-status-selector': StatusSelector,
    },

    props: {
      workId: {
        type: Number,
        required: false
      }
    },

    setup(props, context) {
      const state = reactive({
        libraryEntries: [],
        breadcrumbItems: [],
      })

      onMounted(async () => {
        const [libraryEntries] = await Promise.all([
          new FetchLibraryEntriesQuery().execute(),
        ])
        state.libraryEntries = libraryEntries
        console.log('state.libraryEntries: ', state.libraryEntries)
        state.breadcrumbItems = [
          {
            href: '/',
            text: context.root.$t('noun.home')
          },
          {
            text: context.root.$t('verb.track'),
            current: true
          }
        ]
      })

      return {
        AnnConfig: window.AnnConfig,
        state,
      }
    }
  })
</script>
