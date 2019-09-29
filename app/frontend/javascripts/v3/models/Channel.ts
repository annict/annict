import { ApplicationModel } from './ApplicationModel'
import { Work } from './Work'

export class Channel extends ApplicationModel {
  private annictId: number
  private name: string
  private programs: []

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.name = node.name
  }

  public setProgramsOfWork(work: Work) {
    this.programs = work.programs.filter(program => {
      return program.channel.annictId === this.annictId
    })
  }
}
