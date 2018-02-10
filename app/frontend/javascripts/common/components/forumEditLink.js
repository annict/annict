import eventHub from '../../common/eventHub'

export default {
  template: '#t-forum-edit-link',

  props: {
    userId: {
      type: Number,
      required: true,
    },
    path: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      appData: {},
      appLoaded: false,
    }
  },

  methods: {
    isEditable: function() {
      return this.userId === (this.appData.currentUser && this.appData.currentUser.id)
    },
  },

  mounted() {
    eventHub.$on('app:loaded', () => {
      this.appData = this.$parent.appData
      this.appLoaded = true
    })
  },
}
