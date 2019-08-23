import { ApplicationModel } from './ApplicationModel'

export class Season extends ApplicationModel {
  public constructor(node) {
    super()
    this.slug = node.seasonSlug
    this.name = node.seasonName
    this.localName = node.localSeasonName
    this.year = node.seasonYear
  }

  public isLater() {
    return !this.year && !this.name
  }
}
