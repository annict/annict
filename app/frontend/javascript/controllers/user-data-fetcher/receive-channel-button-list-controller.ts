import BaseController from './base-controller'

export default class extends BaseController {
  static values = {
    url: String,
    channelIds: Array,
  }

  channelIdsValue!: string[];

  replacementSelectors = this.channelIdsValue.map(channelId => `#receive-channel-button-list-${channelId}`)
}
