import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

const NO_SELECT = 'no_select';

export default class extends Controller {
  static targets = ['kind'];
  static values = { animeId: Number, pageCategory: String }

  animeIdValue!: number;
  kindTarget!: HTMLSelectElement;
  statusKinds!: { [animeId: number]: string };
  pageCategoryValue!: string;
  prevStatusKind!: string;

  initialize() {
    this.element.classList.add('c-spinner');

    document.addEventListener('component-value-fetcher:status-selector:fetched', (event: any) => {
      this.statusKinds = event.detail;
      this.kindTarget.value = this.prevStatusKind = this.currentStatusKind;

      if (this.currentStatusKind === NO_SELECT) {
        this.element.classList.add('unselected');
      }

      this.element.classList.remove('c-spinner');
    });
  }

  get currentStatusKind() {
    const statusKind = this.statusKinds[this.animeIdValue]

    if (!statusKind) {
      return NO_SELECT;
    }

    return statusKind;
  }

  resetKind() {
    this.kindTarget.value = NO_SELECT;
  }

  change() {
    if (this.kindTarget.value !== this.prevStatusKind) {
      this.element.classList.add('c-spinner');

      axios
        .post(`/api/internal/works/${this.animeIdValue}/statuses/select`, {
          status_kind: this.kindTarget.value,
          page_category: this.pageCategoryValue,
        })
        .then(() => {
          this.element.classList.remove('unselected');
        })
        .catch(() => {
          ($('.c-sign-up-modal') as any).modal('show');
          this.resetKind();
        })
        .then(() => {
          this.element.classList.remove('c-spinner');
        });
    }
  }
}
