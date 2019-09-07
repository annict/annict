import { ApplicationModel } from './ApplicationModel'
import { User } from './User'
import { Record } from './Record'

export class WorkRecord extends ApplicationModel {
  private id: string
  private ratingAnimationState: string
  private ratingMusicState: string
  private ratingStoryState: string
  private ratingCharacterState: string
  private ratingOverallState: string
  private body: string
  private bodyHtml: string
  private likesCount: number
  private createdAt: DateTime
  private modifiedAt: DateTime
  private viewerDidLike: boolean
  private user: User
  private record: Record

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
