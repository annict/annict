import { ApplicationModel, Work } from '../models'

export class SeriesWork extends ApplicationModel {
  public localSummary: string
  public work: Work

  public constructor(edge) {
    super()
    this.localSummary = edge.localSummary
    this.work = new Work(edge.node)
  }
}
