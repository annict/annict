import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

const NO_SELECT = 'no_select';

export default class extends Controller {
  static targets = ['kind'];
  static values = { animeId: Number, initKind: String, pageCategory: String }

  animeIdValue!: number;
  initKindValue!: string;
  kindTarget!: HTMLSelectElement;
  libraryEntries!: { work_id: number; status_kind: string }[];
  pageCategoryValue!: string;
  prevStatusKind!: string;

  initialize() {
    if (this.initKindValue !== '') {
      this.kindTarget.value = this.initKindValue
      return
    }

    this.element.classList.add('c-spinner');

    document.addEventListener('user-data-fetcher:library-entries:fetched', (event: any) => {
      this.libraryEntries = event.detail;
      this.kindTarget.value = this.prevStatusKind = this.currentStatusKind;

      if (this.currentStatusKind === NO_SELECT) {
        this.element.classList.add('unselected');
      }

      this.element.classList.remove('c-spinner');
    });
  }

  get currentStatusKind() {
    if (!this.libraryEntries.length) {
      return NO_SELECT;
    }

    const status = this.libraryEntries.filter((entry) => {
      return entry.work_id === this.animeIdValue;
    })[0];

    if (!status) {
      return NO_SELECT;
    }

    return status.status_kind;
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
