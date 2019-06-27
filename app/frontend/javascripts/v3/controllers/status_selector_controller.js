import { Controller } from 'stimulus'
import $ from 'jquery'

export default class extends Controller {
  static targets = ['status']

  change(_event) {
    this.element.classList.add('c-spinner')

    $.ajax({
      method: 'PATCH',
      url: '/api/internal/v3/me/status',
      data: {
        work_id: this.data.get('workId'),
        status_kind: this.statusTarget.value,
      },
    }).done(_data => {
      this.element.classList.remove('c-spinner')
    })
  }

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
