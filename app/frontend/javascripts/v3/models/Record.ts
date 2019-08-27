import { ApplicationModel } from './ApplicationModel'

export class Record extends ApplicationModel {
  private annictId: number
  private pageViewsCount: number

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.pageViewsCount = node.pageViewsCount
  }
}
