import { Controller } from 'stimulus';

export default class extends Controller {
  episodeId!: string;

  initialize() {
    this.episodeId = this.data.get('episodeId') ?? '';

    document.addEventListener('trackable-episode-table-row:inactive', ({ detail: { episodeId } }: any) => {
      this.inactive(episodeId);
    });
  }

  inactive(episodeId: string) {
    if (this.episodeId !== episodeId) {
      return
    }

    this.element.classList.add('table-secondary');
  }
}
