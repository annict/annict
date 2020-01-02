<template>
  <transition name="app" mode="out-in">
    <div v-if="state.work" key="content">
      <ann-navbar></ann-navbar>
      <div class="container p-3" data-turbolinks="false">
        <ann-breadcrumb :items="state.breadcrumbItems" class="mb-3"></ann-breadcrumb>
        hello
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
  import StatusSelector from '../components/StatusSelector.vue'

  import { FetchVodChannelsQuery, FetchWorkQuery } from '../queries'

  export default createComponent({
    components: {
      'ann-breadcrumb': Breadcrumb,
      'ann-empty': Empty,
      'ann-footer': Footer,
      'ann-navbar': NavBar,
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
        state.libraryEntries = new FetchLibraryEntriesQuery().execute()
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
