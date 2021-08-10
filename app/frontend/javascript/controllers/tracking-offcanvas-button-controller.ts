import Offcanvas from 'bootstrap/js/dist/offcanvas';
import { Controller } from 'stimulus';

export default class extends Controller {
  static values = { animeId: Number };

  animeIdValue!: number;
  framePath!: string;
  statusKinds!: { [key: number]: string };
  offcanvasElm!: HTMLElement | null;
  offcanvasFrameElm!: HTMLElement | null;

  initialize() {
    this.framePath = this.data.get('framePath') ?? '/500.html';
    this.offcanvasElm = document.querySelector('.c-tracking-offcanvas');
    this.offcanvasFrameElm = this.offcanvasElm?.querySelector('#c-tracking-offcanvas-frame') ?? null;

    document.addEventListener('component-value-fetcher:status-select-dropdown:fetched', (event: any) => {
      if (this.statusKinds) {
        return;
      }

      this.statusKinds = event.detail;

      if (!this.statusKinds[this.animeIdValue]) {
        this.element.setAttribute('disabled', 'true');
      }
    });

    document.addEventListener('tracking-offcanvas-button:enabled', (event: any) => {
      const { animeId } = event.detail;

      if (this.animeIdValue === animeId) {
        this.element.removeAttribute('disabled');
      }
    });
  }

  open() {
    if (this.offcanvasElm && this.offcanvasFrameElm) {
      this.offcanvasFrameElm.setAttribute('src', this.framePath);
      this.offcanvasFrameElm.dataset.reloadableUrlValue = this.framePath;

      new Offcanvas(this.offcanvasElm).show(this.offcanvasElm);
    }
  }
}
