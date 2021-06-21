import Modal from 'bootstrap/js/dist/modal';
import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

import fetcher from '../utils/fetcher';

const NO_SELECT = 'no_select';

export default class extends Controller {
  static classes = ['selected'];
  static targets = ['kind'];
  static values = { animeId: Number, pageCategory: String };

  animeIdValue!: number;
  selectedClass!: string;
  kindTarget!: HTMLSelectElement;
  statusKinds!: { [animeId: number]: string };
  pageCategoryValue!: string;
  prevStatusKind!: string;

  initialize() {
    this.element.classList.add('c-spinner');

    document.addEventListener('component-value-fetcher:status-selector:fetched', (event: any) => {
      this.statusKinds = event.detail;
      this.kindTarget.value = this.prevStatusKind = this.currentStatusKind;

      if (this.currentStatusKind !== NO_SELECT) {
        this.element.classList.add(this.selectedClass);
      }

      this.element.classList.remove('c-spinner');
    });
  }

  get currentStatusKind() {
    const statusKind = this.statusKinds[this.animeIdValue];

    if (!statusKind) {
      return NO_SELECT;
    }

    return statusKind;
  }

  resetKind() {
    this.kindTarget.value = NO_SELECT;
  }

  async change() {
    if (this.kindTarget.value !== this.prevStatusKind) {
      this.element.classList.add('c-spinner');

      try {
        await fetcher.post(`/api/internal/works/${this.animeIdValue}/statuses/select`, {
          status_kind: this.kindTarget.value,
          page_category: this.pageCategoryValue,
        });

        if (this.kindTarget.value === NO_SELECT) {
          this.element.classList.remove(this.selectedClass);
        }
      } catch (err) {
        if (err.response.status === 401) {
          new Modal('.c-sign-up-modal').show();
        }

        this.resetKind();
      } finally {
        this.element.classList.remove('c-spinner');
      }
    }
  }
}
