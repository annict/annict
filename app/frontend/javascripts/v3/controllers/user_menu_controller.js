import { Controller } from 'stimulus'
import $ from 'jquery'

export default class extends Controller {
  connect() {
    $.ajax({
      method: 'GET',
      url: '/api/internal/v3/me/user_menu',
    }).done(data => {
      this.element.innerHTML = data
    })
  }
}
