import { ApplicationModel } from './ApplicationModel'
import {Channel, Episode, Work} from '../models'

export class Slot extends ApplicationModel {
  public channel: Channel
  public work: Work
  public episode: Episode
  public startedAt: string

  public constructor(node) {
    super()
    this.channel = new Channel(node.channel)
    this.work = new Work(node.work)
    this.episode = new Episode(node.episode)
    this.startedAt = node.startedAt
  }
}
