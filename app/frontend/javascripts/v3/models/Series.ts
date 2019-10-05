import { ApplicationModel, SeriesWork } from '../models'

export class Series extends ApplicationModel {
  public localName: string
  public seriesWorks: SeriesWork[]

  public constructor(node) {
    super()
    this.localName = node.localName
  }

  public setSeriesWorks(edges) {
    this.seriesWorks = edges.map(edge => {
      return new SeriesWork(edge)
    })
  }
}
