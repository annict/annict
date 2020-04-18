import { ApplicationModel } from './ApplicationModel'

export class Record extends ApplicationModel {
  public annictId: number
  public pageViewsCount: number

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.pageViewsCount = node.pageViewsCount
  }
}
