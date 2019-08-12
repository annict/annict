import { ApplicationModel } from './ApplicationModel'

export class Episode extends ApplicationModel{
  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.numberText = node.numberText
    this.title = node.title
  }
}
