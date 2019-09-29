import { ApplicationModel } from './ApplicationModel'

export class Person extends ApplicationModel {
  public annictId: number
  public name: string
  public nameEn: string

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.name = node.name
    this.nameEn = node.nameEn
  }
}
