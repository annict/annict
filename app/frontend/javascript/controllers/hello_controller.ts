import { Controller } from 'stimulus'

export default class extends Controller {
  static targets = ['name']

  readonly nameTarget!: HTMLInputElement

  connect(): void {
    console.log("hello, Stimulus!", this.element)
  }

  greet(): void {
    console.log(`Hello, ${this.name}!`)
  }

  get name() {
    return this.nameTarget.value
  }
}
