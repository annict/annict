import Modal from 'bootstrap/js/dist/modal';
import { Controller } from '@hotwired/stimulus';

import fetcher from '../utils/fetcher';

export default class extends Controller {
  static classes = ['default', 'following'];
  static values = {
    userId: Number,
    defaultText: String,
    followingText: String,
  };

  defaultClass!: string;
  defaultTextValue!: string;
  followingClass!: string;
  followingTextValue!: string;
  isLoading!: boolean;
  isFollowing!: boolean;
  userIdValue!: number;

  initialize() {
    this.startLoading();

    document.addEventListener('component-value-fetcher:follow-button:fetched', (event: any) => {
      const userIds = event.detail;

      this.isFollowing = userIds.includes(this.userIdValue);
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
    if (this.isFollowing) {
      this.element.classList.remove(this.defaultClass);
      this.element.classList.add(this.followingClass);
      this.element.innerHTML = `<i class="far fa-check me-1"></i>${this.followingTextValue}`;
    } else {
      this.element.classList.add(this.defaultClass);
      this.element.classList.remove(this.followingClass);
      this.element.innerHTML = `<i class="far fa-plus me-1"></i>${this.defaultTextValue}`;
    }
  }

  async toggle() {
    if (this.isLoading) {
      return;
    }

    this.startLoading();

    try {
      if (this.isFollowing) {
        await fetcher.delete('/api/internal/follow', {
          user_id: this.userIdValue,
        });
        this.isFollowing = false;
      } else {
        await fetcher.post('/api/internal/follow', {
          user_id: this.userIdValue,
        });
        this.isFollowing = true;
      }

      this.render();
    } catch (err) {
      if (err.response?.status === 401) {
        new Modal('.c-sign-up-modal').show();
      }
    } finally {
      this.endLoading();
    }
  }
}
