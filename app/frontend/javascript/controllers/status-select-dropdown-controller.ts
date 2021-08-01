import Modal from 'bootstrap/js/dist/modal';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';
import fetcher from '../utils/fetcher';

enum STATUS_KIND {
  NO_STATUS = 'no_status',
  COMPLETED = 'completed',
  DROPPED = 'dropped',
  ON_HOLD = 'on_hold',
  PLAN_TO_WATCH = 'plan_to_watch',
  WATCHING = 'watching',
}

export default class extends Controller {
  static targets = ['button', 'kind'];
  static values = { animeId: Number, kindIcons: Object, pageCategory: String };

  animeIdValue!: number;
  buttonTarget!: HTMLButtonElement;
  currentStatusKind!: STATUS_KIND;
  kindIconsValue!: any;
  statusKinds!: { [animeId: number]: STATUS_KIND };
  pageCategoryValue!: string;
  prevStatusKind!: STATUS_KIND;

  initialize() {
    this.startLoading();

    document.addEventListener('component-value-fetcher:status-select-dropdown:fetched', (event: any) => {
      if (this.currentStatusKind) {
        return;
      }

      const statusKinds = event.detail;
      this.setCurrentStatusKindFromStatusKinds(statusKinds);
      this.prevStatusKind = this.currentStatusKind;
      this.render();
    });
  }

  setCurrentStatusKindFromStatusKinds(statusKinds: STATUS_KIND[]) {
    const statusKind = statusKinds[this.animeIdValue];

    if (!statusKind) {
      this.currentStatusKind = STATUS_KIND.NO_STATUS;
    } else {
      this.currentStatusKind = statusKind;
    }
  }

  startLoading() {
    this.element.classList.add('c-spinner');
  }

  stopLoading() {
    this.element.classList.remove('c-spinner');
  }

  resetKind() {
    this.currentStatusKind = STATUS_KIND.NO_STATUS;
    this.render();
  }

  render() {
    this.buttonTarget.innerHTML = `<i class="fas fa-${this.kindIconsValue[this.currentStatusKind]}">`;

    if (this.currentStatusKind === STATUS_KIND.NO_STATUS) {
      this.buttonTarget.className = `btn dropdown-toggle u-btn-outline-status`;
    } else {
      this.buttonTarget.className = `btn dropdown-toggle u-bg-${this.currentStatusKind.replace(/_/g, '-')} text-white`;
    }

    this.stopLoading();
  }

  async change(event: any) {
    const selectedStatusKind = event.currentTarget.dataset.statusKind;

    if (selectedStatusKind !== this.currentStatusKind) {
      this.startLoading();

      try {
        await fetcher.post(`/api/internal/animes/${this.animeIdValue}/status_select`, {
          status_kind: selectedStatusKind,
          page_category: this.pageCategoryValue,
        });

        this.currentStatusKind = selectedStatusKind;
        this.render();
        new EventDispatcher('tracking-modal-button:enabled', { animeId: this.animeIdValue }).dispatch();
      } catch (err) {
        if (err.response.status === 401) {
          new Modal('.c-sign-up-modal').show();
        }

        this.resetKind();
      } finally {
        this.stopLoading();
      }
    }
  }
}
