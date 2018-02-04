import $ from 'jquery'

export default {
  loadAppData() {
    return $.ajax({
      method: 'GET',
      url: '/api/internal/app_data',
    })
  },

  loadPageData() {
    if (!gon.page.params) {
      return
    }

    return $.ajax({
      method: 'GET',
      url: '/api/internal/page_data',
      data: {
        page_category: gon.page.category,
        page_params: gon.page.params,
      },
    })
  },
}
