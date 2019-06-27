import { Controller } from 'stimulus'
import $ from 'jquery'

export default class extends Controller {
  static targets = ['status']

  connect() {
    $.ajax({
      method: 'GET',
      url: '/api/internal/v3/me/status',
      data: {
        work_id: this.data.get('workId'),
      },
    }).done(data => {
      this.statusTarget.value = data.status.toLowerCase()
    })
  }
}
