import { ApplicationModel } from './ApplicationModel'

export class Trailer extends ApplicationModel {
  public internalImageUrl: string
  public title: string
  public url: string

  public constructor(node) {
    super()
    this.internalImageUrl = node.internalImageUrl
    this.title = node.title
    this.url = node.url
  }
}
