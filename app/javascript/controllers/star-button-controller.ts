import Modal from 'bootstrap/js/dist/modal';
import { Controller } from '@hotwired/stimulus';

import fetcher from '../utils/fetcher';

type Star = {
  starrable_type: string;
  starrable_id: number;
};

export default class extends Controller {
  static classes = ['default', 'starred'];
  static values = {
    starrableId: Number,
    starrableType: String,
  };

  defaultClass!: string;
  starredClass!: string;
  isLoading!: boolean;
  hasStarred!: boolean;
  starrableIdValue!: number;
  starrableTypeValue!: string;

  initialize() {
    this.startLoading();

    document.addEventListener('component-value-fetcher:star-button:fetched', (event: any) => {
      const stars = event.detail;

      this.hasStarred = !!stars.find(
        (star: Star) => star.starrable_id === this.starrableIdValue && star.starrable_type === this.starrableTypeValue,
      );
      this.render();

      this.endLoading();
    });
  }

  startLoading() {
    this.element.classList.add('c-spinner');
    this.isLoading = true;
  }

  endLoading() {
    this.element.classList.remove('c-spinner');
    this.isLoading = false;
  }

  render() {
    if (this.hasStarred) {
      this.element.classList.remove(this.defaultClass);
      this.element.classList.add(this.starredClass);
      this.element.innerHTML = '<i class="fas fa-star"></i>';
    } else {
      this.element.classList.add(this.defaultClass);
      this.element.classList.remove(this.starredClass);
      this.element.innerHTML = '<i class="far fa-star"></i>';
    }
  }

  async toggle() {
    if (this.isLoading) {
      return;
    }

    this.startLoading();

    try {
      if (this.hasStarred) {
        await fetcher.post('/api/internal/unstars', {
          starrable_type: this.starrableTypeValue,
          starrable_id: this.starrableIdValue,
        });
        this.hasStarred = false;
      } else {
        await fetcher.post('/api/internal/stars', {
          starrable_type: this.starrableTypeValue,
          starrable_id: this.starrableIdValue,
        });
        this.hasStarred = true;
      }

      this.render();
    } catch (err) {
      console.error(err);

      if (err.response?.status === 401) {
        new Modal('.c-sign-up-modal').show();
      }
    } finally {
      this.endLoading();
    }
  }
}
