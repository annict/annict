import $ from 'jquery'
import _ from 'lodash'

export default {
  template: '#t-search-form',

  data() {
    return {
      works: [],
      people: [],
      organizations: [],
      index: -1,
      q: this.initQ,
    }
  },

  props: {
    initQ: String,
  },

  computed: {
    results() {
      _.each(this.works, work => (work.resourceType = 'work'))
      _.each(this.people, person => (person.resourceType = 'person'))
      _.each(this.organizations, org => (org.resourceType = 'organization'))
      _.each(this.characters, char => (char.resourceType = 'character'))
      const results = []
      results.push.apply(results, this.works)
      results.push.apply(results, this.people)
      results.push.apply(results, this.organizations)
      results.push.apply(results, this.characters)
      return results
    },
  },

  methods: {
    resultPath(result) {
      const resourceName = (() => {
        switch (result.resourceType) {
          case 'work':
            return 'works'
          case 'person':
            return 'people'
          case 'organization':
            return 'organizations'
          case 'character':
            return 'characters'
        }
      })()
      return `/${resourceName}/${result.id}`
    },

    onKeyup: _.debounce(function() {
      return $.ajax({
        method: 'GET',
        url: '/api/internal/search',
        data: {
          q: this.q,
        },
      }).done(data => {
        this.works = data.works
        this.people = data.people
        this.organizations = data.organizations
        return (this.characters = data.characters)
      })
    }, 300),

    next() {
      if (this.results.length) {
        this.index += 1
        if (this.index === this.results.length) {
          return (this.index = -1)
        }
      }
    },

    prev() {
      if (this.results.length) {
        this.index -= 1
        if (this.index === -2) {
          return (this.index = this.results.length - 1)
        }
      }
    },

    select(event) {
      if (this.index !== -1) {
        location.href = this.resultPath(this.results[this.index])
        event.preventDefault()
      }
    },

    onMouseover(index) {
      return (this.index = index)
    },

    hideResults() {
      this.works = this.people = this.organizations = this.characters = []
      return $(this.$el)
        .find('input')
        .blur()
    },
  },
}
