import { ApplicationModel } from './ApplicationModel'
import { Channel } from './Channel'

export class Program extends ApplicationModel{
  public annictId: number
  public vodTitleName: string
  public vodTitleCode: string
  public vodTitleUrl: string
  public channel: Channel

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
