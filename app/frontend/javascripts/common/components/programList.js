import $ from 'jquery'
import _ from 'lodash'
import moment from 'moment'

import eventHub from '../../common/eventHub'
import vueLazyLoad from '../../common/vueLazyLoad'
import loadMoreButton from './loadMoreButton'

export default {
  template: '#t-program-list',

  data() {
    return {
      isLoading: false,
      hasNext: true,
      programs: [],
      user: null,
      page: 1,
      sort: gon.currentProgramsSortType,
      sortTypes: gon.programsSortTypes,
    }
  },

  components: {
    'c-load-more-button': loadMoreButton,
  },

  methods: {
    requestData() {
      const data = {
        page: this.page,
        sort: this.sort,
      }
      return data
    },

    initPrograms(programs) {
      return _.each(programs, function(program) {
        program.isBroadcasted = moment().isAfter(program.started_at)
        return (program.record = {
          uid: _.uniqueId(),
          body: '',
          isEditingComment: false,
          isRecorded: false,
          isSaving: false,
          ratingState: null,
          wordCount: 0,
          bodyRows: 1,
        })
      })
    },

    loadMore() {
      if (this.isLoading) {
        return
      }

      this.isLoading = true
      this.page += 1

      return $.ajax({
        method: 'GET',
        url: '/api/internal/user/programs',
        data: this.requestData(),
      }).done(data => {
        this.isLoading = false
        if (data.programs.length > 0) {
          this.hasNext = true
          return this.programs.push.apply(this.programs, this.initPrograms(data.programs))
        } else {
          return (this.hasNext = false)
        }
      })
    },

    reload() {
      return this.updateProgramsSortType(() => (location.href = '/programs'))
    },

    submit(program) {
      if (program.record.isSaving || program.record.isRecorded) {
        return
      }

      program.record.isSaving = true

      return $.ajax({
        method: 'POST',
        url: `/api/internal/episodes/${program.episode.id}/records`,
        data: {
          episode_record: {
            body: program.record.body,
            shared_twitter: this.user.share_record_to_twitter,
            rating_state: program.record.ratingState,
          },
          page_category: gon.page.category,
        },
      })
        .done(function(data) {
          program.record.isSaving = false
          program.record.isRecorded = true
          const msg = gon.I18n['messages.components.program_list.tracked']
          return eventHub.$emit('flash:show', msg)
        })
        .fail(function(data) {
          program.record.isSaving = false
          return eventHub.$emit('flash:show', data.responseJSON.message, 'alert')
        })
    },

    load() {
      this.programs = this.initPrograms(this._pageObject().programs)
      this.hasNext = this.programs.length > 0
      this.user = this._pageObject().user
      return this.$nextTick(() => vueLazyLoad.refresh())
    },

    updateProgramsSortType(callback) {
      return $.ajax({
        method: 'PATCH',
        url: '/api/internal/programs_sort_type',
        data: {
          programs_sort_type: this.sort,
        },
      }).done(callback)
    },

    _pageObject() {
      if (!gon.pageObject) {
        return {}
      }
      return JSON.parse(gon.pageObject)
    },
  },

  mounted() {
    return this.load()
  },
}
