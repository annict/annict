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
  // isWatched!: boolean;
  loadingClass!: string;
  // pageCategory!: string | null;
  // recordIdValue!: number | null;

  initialize() {
    // this.episodeId = Number(this.data.get('episodeId'));
    // this.pageCategory = this.data.get('pageCategory');
  }

  // render() {
  //   if (this.isWatched) {
  //     this.element.classList.add('btn-info');
  //     this.element.classList.remove('btn-outline-info');
  //   } else {
  //     this.element.classList.remove('btn-info');
  //     this.element.classList.add('btn-outline-info');
  //   }
  // }

  // endLoading() {
  //   this.isLoading = false
  //   this.element.classList.remove(this.loadingClass);
  // }

  watch() {
    if (this.isLoading) {
      return;
    }

    const listGroupElm = this.element.closest('.c-tracking-episode-list-group')

    if (!listGroupElm) {
      return
    }

    // const episodeIds = Array.from(listGroupElm.querySelectorAll('.list-group-item')).map((itemElm) => {
    //   if (itemElm instanceof HTMLElement) {
    //     return Number(itemElm.dataset.episodeId)
    //   }
    // })
    // const targetEpisodeIds = episodeIds.slice(episodeIds.indexOf(this.episodeIdValue))

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
    })

    axios
      .post('/api/internal/multiple_episode_records', {
        episode_ids: targetEpisodeIds,
      })
      .then(() => {
        episodeItemElms.forEach(episodeItemElm => {
          episodeItemElm.remove()
        })
        location.reload();
      })
      .catch(() => {
        ($('.c-sign-up-modal') as any).modal('show');
      })
  }
}
