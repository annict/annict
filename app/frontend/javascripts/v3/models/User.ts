import { ApplicationModel } from './ApplicationModel'

export class User extends ApplicationModel {
  private annictId: number
  private username: string
  private name: string
  private avatarUrl: string
  private isSupporter: boolean

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.username = node.username
    this.name = node.name
    this.avatarUrl = node.avatarUrl
    this.isSupporter = node.isSupporter
  }
}
