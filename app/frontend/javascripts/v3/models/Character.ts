import { ApplicationModel } from './ApplicationModel'

export class Character extends ApplicationModel{
  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.name = node.name
    this.nameEn = node.nameEn
  }
}
