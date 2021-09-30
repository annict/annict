import $ from 'jquery';
import axios from 'axios';
import { Controller } from '@hotwired/stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  static classes = ['loading'];
  static values = {
    episodeId: Number,
  };

  episodeIdValue!: number;
  isLoading!: boolean;
  loadingClass!: string;

  reloadList() {
    new EventDispatcher('reloadable--trackable-episode-list:reload').dispatch();
    new EventDispatcher('reloadable--tracking-offcanvas:reload').dispatch();
  }

  startLoading() {
    this.isLoading = true;
    this.element.classList.add(this.loadingClass);
    this.element.setAttribute('disabled', 'true');
  }

  watch() {
    if (this.isLoading) {
      return;
    }

    const listGroupElm = this.element.closest('.c-tracking-episode-list-group');

    if (!listGroupElm) {
      return;
    }

    const episodeItemElms = Array.from(listGroupElm.querySelectorAll('.list-group-item'));
    const clickedEpisodeItemElmIndex = episodeItemElms.findIndex((itemElm) => {
      return Number((itemElm as HTMLElement).dataset.episodeId) === this.episodeIdValue;
    });
    const targetEpisodeIds = episodeItemElms.slice(0, clickedEpisodeItemElmIndex + 1).map((episodeItemElm) => {
      return (episodeItemElm as HTMLElement).dataset.episodeId;
    });

    this.startLoading();

    axios
      .post('/api/internal/multiple_episode_records', {
        episode_ids: targetEpisodeIds,
      })
      .then(() => {
        this.reloadList();
      })
      .catch(() => {
        ($('.c-sign-up-modal') as any).modal('show');
      });
  }
}
