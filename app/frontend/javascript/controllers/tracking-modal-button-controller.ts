import Modal from 'bootstrap/js/dist/modal';
import { Controller } from 'stimulus';

export default class extends Controller {
  framePath!: string;

  initialize() {
    this.framePath = this.data.get('framePath') ?? '/500.html';
  }

  open() {
    const modalElm = document.querySelector('.c-tracking-modal');
    const frameElm = modalElm?.querySelector('#c-tracking-modal-frame');

    if (modalElm && frameElm) {
      frameElm.setAttribute('src', this.framePath);
      (frameElm as HTMLElement).dataset.reloadableUrlValue = this.framePath;

      new Modal(modalElm).show();
    }
  }
}
