import { Controller } from 'stimulus';

export default class extends Controller {
  initialize() {
    const frameElm = this.element.querySelector('turbo-frame');
    const initHtml = frameElm?.innerHTML ?? '';

    this.element.addEventListener('hidden.bs.offcanvas', () => {
      if (frameElm) {
        frameElm.setAttribute('src', '');
        frameElm.innerHTML = initHtml;
      }
    });
  }
}
