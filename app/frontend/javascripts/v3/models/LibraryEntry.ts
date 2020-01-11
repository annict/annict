import { ApplicationModel } from './ApplicationModel'
import {Channel, Work} from '../models'

export class LibraryEntry extends ApplicationModel {
  public work: Work
  public untappedEpisodesCount: number

  public constructor(node) {
    super()
    this.work = null
    this.untappedEpisodesCount = node.untappedEpisodes.totalCount
  }

  public setWork(node) {
    this.work = new Work(node)
  }
}
