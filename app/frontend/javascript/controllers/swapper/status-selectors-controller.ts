import BaseController from './base-controller'

export default class extends BaseController {
  static values = {
    url: String,
    animeIds: Array,
  }

  animeIdsValue!: string[];

  swapSelectors = this.animeIdsValue.map(animeId => `#status-selector-anime-${animeId}`)
}
