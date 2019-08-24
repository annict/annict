import { ApplicationModel } from './ApplicationModel'
import { Channel } from './Channel'

export class Program extends ApplicationModel{
  private annictId: number
  private vodTitleName: string
  private vodTitleCode: string
  private vodTitleUrl: string
  private channel: Channel

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.vodTitleName = node.vodTitleName
    this.vodTitleCode = node.vodTitleCode
    this.vodTitleUrl = node.vodTitleUrl
    this.channel = null
  }

  public setChannel(node) {
    this.channel = new Channel(node)
  }
}
