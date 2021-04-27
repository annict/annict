import BaseController from './base-controller'

export default class extends BaseController {
  static values = {
    url: String,
    channelIds: Array,
  }

  channelIdsValue!: string[];

  swapSelectors = this.channelIdsValue.map(channelId => `#receive-channel-button-list-${channelId}`)
}
