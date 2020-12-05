import axios from 'axios';
import { Controller } from 'stimulus';

import { EventDispatcher } from '../utils/event-dispatcher';

export default class extends Controller {
  static targets = ['episodeRow'];

  isLoading!: boolean;
  episodeRowTargets!: [HTMLElement];

  initialize() {
    this.isLoading = false;
  }

  async trackEpisodes(event: Event) {
    if (this.isLoading) {
      return;
    }

    this.isLoading = true;

    const { episodeId: currentEpisodeId } = (event.currentTarget as HTMLElement).dataset;

    let isCurrentEpisodeIdMatched = false
    const episodeRowElms = this.episodeRowTargets.filter((episodeRowElm) => {
      const { trackableEpisodeTableRowEpisodeId: episodeId, trackableEpisodeTableRowTracked: isTracked } = episodeRowElm.dataset

      if (currentEpisodeId === episodeId) {
        isCurrentEpisodeIdMatched = true
      }

      if (isCurrentEpisodeIdMatched && !isTracked) {
        return true
      }

      return false
    })
    const episodeIds = episodeRowElms.map(episodeRowElm => episodeRowElm.dataset.trackableEpisodeTableRowEpisodeId)
    const trackButtonElms = episodeRowElms.map(episodeRowElm => episodeRowElm.querySelector('.c-bulk-track-episode-button'))

    trackButtonElms.forEach((trackButtonElm) => {
      trackButtonElm?.setAttribute('disabled', '');
      trackButtonElm?.classList.add('c-spinner')
    })

    const res = await axios.
      post(`/api/internal/multiple_episode_records`, { episode_ids: episodeIds }).
      catch((err) => console.error(err))

    if (!res) {
      return
    }

    const jobId = res.data.job_id

    if (await this.isJobProcessed(jobId)) {
      this.isLoading = false;

      episodeRowElms.forEach((episodeRowElm) => {
        const episodeId = episodeRowElm.dataset.trackableEpisodeTableRowEpisodeId
        const trackButtonElm = episodeRowElm.querySelector('.c-bulk-track-episode-button')

        trackButtonElm?.removeAttribute('disabled');
        trackButtonElm?.classList.remove('c-spinner')

        new EventDispatcher('trackable-episode-table-row:inactive', { episodeId: episodeId }).dispatch();
      })
    }
  }

  async isJobProcessed(jobId: string): Promise<boolean> {
    const res = await axios.
      get(`/api/internal/bulk_operations/${jobId}`).
      catch((err) => console.error(err))

    if (!res) {
      return false
    }

    if (res.data.job_id) {
      await new Promise(resolve => setTimeout(resolve, 3000))

      return this.isJobProcessed(jobId)
    }

    return true
  }
}
