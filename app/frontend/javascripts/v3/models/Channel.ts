import { ApplicationModel } from './ApplicationModel'
import { Program, Work } from '../models'

export class Channel extends ApplicationModel {
  public annictId: number
  public name: string
  public programs: Program[]

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
