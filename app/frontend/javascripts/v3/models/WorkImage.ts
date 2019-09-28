import { ApplicationModel } from './ApplicationModel'

export class WorkImage extends ApplicationModel {
  private internalUrl: string

  public constructor(node) {
    super()
    this.internalUrl = node.internalUrl
  }
}
