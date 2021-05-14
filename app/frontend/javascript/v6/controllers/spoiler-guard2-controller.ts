import { Controller } from 'stimulus';

export default class extends Controller {
  static values = { initIsSpoiler: Boolean }

  initIsSpoilerValue!: boolean;
  isSpoiler!: boolean;

  connect() {
    this.isSpoiler = this.initIsSpoilerValue;
    this.render();
  }

  render() {
    if (this.isSpoiler) {
      this.element.classList.add('is-spoiler');
      this.element.classList.remove('is-not-spoiler');
    } else {
      this.element.classList.remove('is-spoiler');
      this.element.classList.add('is-not-spoiler');
    }
  }

  hide() {
    this.isSpoiler = false;
    this.render();
  }
}
