import { ApplicationModel } from './ApplicationModel'

export class Trailer extends ApplicationModel {
  public constructor(node) {
    super()
    this.internalImageUrl = node.internalImageUrl
    this.title = node.title
    this.url = node.url
  }
}
