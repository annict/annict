import $ from 'jquery'
import _ from 'lodash'
import moment from 'moment'

import eventHub from '../../common/eventHub'
import vueLazyLoad from '../../common/vueLazyLoad'
import loadMoreButton from './loadMoreButton'

export default {
  template: '#t-slot-list',

  data() {
    return {
      isLoading: false,
      hasNext: true,
      slots: [],
      user: null,
      page: 1,
      sort: gon.currentSlotsSortType,
      sortTypes: gon.slotsSortTypes,
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

    initSlots(slots) {
      return _.each(slots, function(slot) {
        slot.isBroadcasted = moment().isAfter(slot.started_at)
        return (slot.record = {
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
        url: '/api/internal/user/slots',
        data: this.requestData(),
      }).done(data => {
        this.isLoading = false
        if (data.slots.length > 0) {
          this.hasNext = true
          return this.slots.push.apply(this.slots, this.initSlots(data.slots))
        } else {
          return (this.hasNext = false)
        }
      })
    },

    reload() {
      return this.updateSlotsSortType(() => (location.href = '/slots'))
    },

    submit(slot) {
      if (slot.record.isSaving || slot.record.isRecorded) {
        return
      }

      slot.record.isSaving = true

      return $.ajax({
        method: 'POST',
        url: `/api/internal/episodes/${slot.episode.id}/records`,
        data: {
          episode_record: {
            body: slot.record.body,
            shared_twitter: this.user.share_record_to_twitter,
            rating_state: slot.record.ratingState,
          },
          page_category: gon.page.category,
        },
      })
        .done(function(data) {
          slot.record.isSaving = false
          slot.record.isRecorded = true
          const msg = gon.I18n['messages.components.slot_list.tracked']
          return eventHub.$emit('flash:show', msg)
        })
        .fail(function(data) {
          slot.record.isSaving = false
          return eventHub.$emit('flash:show', data.responseJSON.message, 'alert')
        })
    },

    load() {
      this.slots = this.initSlots(this._pageObject().slots)
      this.hasNext = this.slots.length > 0
      this.user = this._pageObject().user
      return this.$nextTick(() => vueLazyLoad.refresh())
    },

    updateSlotsSortType(callback) {
      return $.ajax({
        method: 'PATCH',
        url: '/api/internal/slots_sort_type',
        data: {
          slots_sort_type: this.sort,
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
