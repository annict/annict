import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

const NO_SELECT = 'no_select';

export default class extends Controller {
  static targets = ['kind'];

  kindTarget!: HTMLSelectElement;
  libraryEntries!: { work_id: number; status_kind: string }[];
  prevStatusKind!: string;
  workId!: number;
  pageCategory!: string;

  initialize() {
    this.workId = Number(this.data.get('workId'));
    this.pageCategory = this.data.get('pageCategory') || '';

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
      return entry.work_id === this.workId;
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
        .post(`/api/internal/works/${this.workId}/statuses/select`, {
          status_kind: this.kindTarget.value,
          page_category: this.pageCategory,
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
