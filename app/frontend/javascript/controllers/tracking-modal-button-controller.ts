import Modal from 'bootstrap/js/dist/modal';
import { Controller } from 'stimulus';

export default class extends Controller {
  static values = { animeId: Number };

  animeIdValue!: number;
  framePath!: string;
  statusKinds!: { [key: number]: string };

  initialize() {
    this.framePath = this.data.get('framePath') ?? '/500.html';

    document.addEventListener('component-value-fetcher:status-select-dropdown:fetched', (event: any) => {
      if (this.statusKinds) {
        return;
      }

      this.statusKinds = event.detail;

      if (!this.statusKinds[this.animeIdValue]) {
        this.element.setAttribute('disabled', 'true');
      }
    });

    document.addEventListener('tracking-modal-button:enabled', (event: any) => {
      const { animeId } = event.detail;

      if (this.animeIdValue === animeId) {
        this.element.removeAttribute('disabled');
      }
    });
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
