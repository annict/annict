import Modal from 'bootstrap/js/dist/modal';
import { Controller } from '@hotwired/stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';
import fetcher from '../utils/fetcher';

export default class extends Controller {
  static classes = ['default', 'muted'];
  static values = {
    userId: Number,
    defaultText: String,
    mutedText: String,
  };

  defaultClass!: string;
  defaultTextValue!: string;
  mutedClass!: string;
  mutedTextValue!: string;
  isLoading!: boolean;
  isMuted!: boolean;
  userIdValue!: number;

  initialize() {
    this.startLoading();

    document.addEventListener('component-value-fetcher:mute-user-button:fetched', (event: any) => {
      const userIds = event.detail;

      this.isMuted = userIds.includes(this.userIdValue);
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
    if (this.isMuted) {
      this.element.classList.remove(this.defaultClass);
      this.element.classList.add(this.mutedClass);
      this.element.innerHTML = this.mutedTextValue;
    } else {
      this.element.classList.add(this.defaultClass);
      this.element.classList.remove(this.mutedClass);
      this.element.innerHTML = this.defaultTextValue;
    }
  }

  async toggle() {
    if (this.isLoading) {
      return;
    }

    this.startLoading();

    try {
      let res: any;

      if (this.isMuted) {
        res = await fetcher.delete('/api/internal/mute_user', {
          user_id: this.userIdValue,
        });
      } else {
        res = await fetcher.post('/api/internal/mute_user', {
          user_id: this.userIdValue,
        });
      }

      if (res?.flash) {
        new EventDispatcher('flash:show', res.flash).dispatch();
      }

      this.isMuted = !this.isMuted;
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
