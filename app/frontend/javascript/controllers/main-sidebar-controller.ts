import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  backgroundElm!: Element | null;
  contentElm!: Element | null;

  initialize() {
    this.backgroundElm = this.element.querySelector('.c-main-sidebar__background');
    this.contentElm = this.element.querySelector('.c-main-sidebar__content');

    document.addEventListener('main-sidebar:show', (event: any) => {
      this.show();
    });
  }

  show() {
    if (this.backgroundElm && this.contentElm) {
      this.backgroundElm.classList.add('active');
      this.contentElm.classList.add('active');
    }
  }

  hide() {
    if (this.backgroundElm && this.contentElm) {
      this.backgroundElm.classList.remove('active');
      this.contentElm.classList.remove('active');
    }
  }
}
