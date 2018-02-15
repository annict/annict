import _ from 'lodash'

import eventHub from '../eventHub'

const DISPLAY_USERS_LIMIT = 12

export default {
  template: '#t-work-friends',

  data() {
    return {
      isSignedIn: window.gon.user.isSignedIn,
      showAll: false,
      works: [],
      workListData: window.gon.workListData ? JSON.parse(window.gon.workListData) : {},
    }
  },

  props: {
    workId: {
      type: Number,
      required: true,
    },
  },

  computed: {
    allUsers() {
      if (!this.works.length) {
        return []
      }
      const data = _.find(this.works, work => {
        return work.id === this.workId
      })
      return data.users
    },

    users() {
      if (this.showAll) {
        return this.allUsers
      }
      return _.take(this.allUsers, DISPLAY_USERS_LIMIT)
    },

    isMoreUsers() {
      return !this.showAll && this.allUsers.length > DISPLAY_USERS_LIMIT
    },
  },

  methods: {
    more() {
      return (this.showAll = true)
    },
  },

  mounted() {
    if (!this.isSignedIn) {
      return
    }
    return (this.works = this.workListData.works)
  },
}
