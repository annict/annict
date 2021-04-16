import $ from 'jquery';
import { Controller } from 'stimulus';

export default class extends Controller {
  framePath!: string;

  initialize() {
    this.framePath = this.data.get('framePath') ?? '/500.html';
  }

  open() {
    const frameElm = document.getElementById('c-tracking-modal__frame')
    frameElm?.setAttribute('src', this.framePath)
    $('#c-tracking-modal__frame').modal()
  }
}
