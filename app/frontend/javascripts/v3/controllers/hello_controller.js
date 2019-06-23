import { Controller } from 'stimulus'

export default class extends Controller {
  static targets = ['name']

  initialize() {
    if (!document.documentElement.hasAttribute("data-turbolinks-preview")) {
      console.log("stimulus:initialized")
    }
  }

  connect() {
    console.log("stimulus:connected", this.element)
  }

  greet() {
    console.log(`Hello, ${this.nameTarget.value}!`)
  }
}
