import { Controller } from 'stimulus'
import $ from 'jquery'

export default class extends Controller {
  connect() {
    console.log('stimulus:connected', this.element)
    $.ajax({
      method: 'GET',
      url: '/api/internal/v3/me/user_menu',
    }).done(data => {
      console.log('data: ', data)
      this.element.innerHTML = data
    })
  }
}
