import { Controller } from '@hotwired/stimulus';

export default class extends Controller {
  static targets = ['content', 'readMoreButton', 'readMoreBackground'];

  contentTarget!: HTMLElement;
  readMoreBackgroundTarget!: HTMLElement;
  readMoreButtonTarget!: HTMLElement;
  height!: number;

  initialize() {
    if (!this.data.get('height')) {
      return;
    }

    this.height = Number(this.data.get('height'));

    if (this.height < this.contentTarget.scrollHeight) {
      this.contentTarget.style.cssText = `height: ${this.height}px; overflow-y: hidden;`;
      this.readMoreBackgroundTarget.style.cssText = `height: ${this.height}px;`;

      this.readMoreBackgroundTarget.classList.remove('d-none');
      this.readMoreButtonTarget.classList.remove('d-none');
    }
  }

  readMore() {
    this.contentTarget.style.removeProperty('height');
    this.contentTarget.style.removeProperty('overflow-y');
    this.readMoreBackgroundTarget.classList.add('d-none');
    this.readMoreButtonTarget.classList.add('d-none');
  }
}
