import { ApplicationModel } from './ApplicationModel'
import { User } from './User'
import { Record } from './Record'

export class WorkRecord extends ApplicationModel {
  public id: string
  public ratingAnimationState: string
  public ratingMusicState: string
  public ratingStoryState: string
  public ratingCharacterState: string
  public ratingOverallState: string
  public body: string
  public bodyHtml: string
  public likesCount: number
  public createdAt: string
  public modifiedAt: string
  public viewerDidLike: boolean
  public user: User
  public record: Record

  public constructor(node) {
    super()
    this.id = node.id
    this.ratingAnimationState = node.ratingAnimationState
    this.ratingMusicState = node.ratingMusicState
    this.ratingStoryState = node.ratingStoryState
    this.ratingCharacterState = node.ratingCharacterState
    this.ratingOverallState = node.ratingOverallState
    this.body = node.body
    this.bodyHtml = node.bodyHtml
    this.likesCount = node.likesCount
    this.createdAt = node.createdAt
    this.modifiedAt = node.modifiedAt
    this.viewerDidLike = node.viewerDidLike
  }

  public setUser(node) {
    this.user = new User(node)
  }

  public setRecord(node) {
    this.record = new Record(node)
  }
}
