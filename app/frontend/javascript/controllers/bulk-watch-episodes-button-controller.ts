import $ from 'jquery';
import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  static classes = ['loading']
  static values = {
    episodeId: Number
  }

  episodeIdValue!: number;
  isLoading!: boolean;
  loadingClass!: string;

  reloadList() {
    new EventDispatcher('reloadable-frame-trackable-episode-list:reload').dispatch();
    new EventDispatcher('reloadable-frame-tracking-modal:reload').dispatch();
  }

  watch() {
    if (this.isLoading) {
      return;
    }

    const listGroupElm = this.element.closest('.c-tracking-episode-list-group')

    if (!listGroupElm) {
      return
    }

    let isCurrentEpisodeIdMatched = false
    const episodeItemElms = Array.from(listGroupElm.querySelectorAll('.list-group-item')).filter((itemElm) => {
      if (!(itemElm instanceof HTMLElement)) {
        return
      }

      if (Number(itemElm.dataset.episodeId) === this.episodeIdValue) {
        isCurrentEpisodeIdMatched = true
      }

      if (isCurrentEpisodeIdMatched) {
        return true
      }

      return false
    })
    const targetEpisodeIds = episodeItemElms.map(episodeItemElm => {
      if (episodeItemElm instanceof HTMLElement) return episodeItemElm.dataset.episodeId
    })
    const watchButtonElms = episodeItemElms.map(episodeItemElm => episodeItemElm.querySelector('.c-bulk-watch-episodes-button'))

    this.isLoading = true
    watchButtonElms.forEach(watchButtonElm => {
      watchButtonElm?.classList?.add(this.loadingClass);
      watchButtonElm?.setAttribute('disabled', "true");
    })

    axios
      .post('/api/internal/multiple_episode_records', {
        episode_ids: targetEpisodeIds,
      })
      .then(() => {
        this.reloadList();
      })
      .catch(() => {
        ($('.c-sign-up-modal') as any).modal('show');
      })
  }
}
