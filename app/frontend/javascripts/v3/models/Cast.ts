import { ApplicationModel, Character, Person } from '../models'

export class Cast extends ApplicationModel {
  public constructor(node) {
    super()
    this.name = node.name
    this.nameEn = node.nameEn
    this.localAccuratedName = node.localAccuratedName
    this.character = {}
    this.person = {}
  }

  public setCharacter(node) {
    this.character = new Character(node)
  }

  public setPerson(node) {
    this.person = new Person(node)
  }
}
