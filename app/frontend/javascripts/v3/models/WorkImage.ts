import { ApplicationModel } from './ApplicationModel'

export class WorkImage extends ApplicationModel {
  public internalUrl: string

  public constructor(node) {
    super()
    this.internalUrl = node?.internalUrl
  }
}
