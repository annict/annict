import axios from 'axios';
import { Controller } from '@hotwired/stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  static classes = ['loading'];
  static values = {
    episodeId: Number,
    pageCategory: String,
  };

  episodeIdValue!: number;
  isLoading!: boolean;
  loadingClass!: string;
  pageCategoryValue!: string;

  startLoading() {
    this.isLoading = true;
    this.element.classList.add(this.loadingClass);
    this.element.setAttribute('disabled', 'true');
  }

  endLoading() {
    this.isLoading = false;
    this.element.classList.remove(this.loadingClass);
  }

  reloadList() {
    new EventDispatcher('reloadable--trackable-episode-list:reload').dispatch();
    new EventDispatcher('reloadable--tracking-offcanvas:reload').dispatch();
  }

  watch() {
    if (this.isLoading) {
      return;
    }

    this.startLoading();

    axios
      .post('/api/internal/episode_records', {
        episode_id: this.episodeIdValue,
        page_category: this.pageCategoryValue,
      })
      .then((res: any) => {
        this.endLoading();
        this.reloadList();
      })
      .catch(() => {
        ($('.c-sign-up-modal') as any).modal('show');
      });
  }
}
