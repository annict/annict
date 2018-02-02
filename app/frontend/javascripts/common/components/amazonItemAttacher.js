import $ from 'jquery'

import eventHub from '../../common/eventHub'

export default {
  template: '#t-amazon-item-attacher',

  props: {
    resourceType: {
      type: String,
      required: true,
    },

    resourceId: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      keyword: '',
      page: 1,
      items: [],
      isLoading: false,
    }
  },

  methods: {
    search(page) {
      if (!this.keyword) {
        return
      }

      $(window).scrollTop(0)

      this.isLoading = true

      return $.ajax({
        method: 'GET',
        url: '/api/internal/amazon/search',
        data: {
          keyword: this.keyword,
          resource_type: this.resourceType,
          resource_id: this.resourceId,
          page,
        },
      })
        .done(data => {
          this.items = data.items
          this.page = page
          return (this.totalPages = data.total_pages)
        })
        .fail(function() {
          const message = gon.I18n['messages._components.amazon_item_attacher.error']
          return eventHub.$emit('flash:show', message, 'alert')
        })
        .always(() => {
          return (this.isLoading = false)
        })
    },

    add(item) {
      if (item.added_to_resource) {
        return
      }

      item.isLoading = true

      return $.ajax({
        method: 'POST',
        url: '/api/internal/items',
        data: {
          resource_type: this.resourceType,
          resource_id: this.resourceId,
          asin: item.asin,
          page_category: gon.basic.pageCategory,
        },
      }).done(function() {
        item.isLoading = false
        return (item.added_to_resource = true)
      })
    },
  },
}
