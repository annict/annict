import { ApplicationModel } from './ApplicationModel'

export class User extends ApplicationModel {
  public annictId: number
  public username: string
  public name: string
  public avatarUrl: string
  public isSupporter: boolean

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.username = node.username
    this.name = node.name
    this.avatarUrl = node.avatarUrl
    this.isSupporter = node.isSupporter
  }
}
