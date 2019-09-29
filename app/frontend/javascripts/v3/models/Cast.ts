import { ApplicationModel, Character, Person } from '../models'

export class Cast extends ApplicationModel {
  public name: string
  public nameEn: string
  public localAccuratedName: string
  public character: Character
  public person: Person

  public constructor(node) {
    super()
    this.name = node.name
    this.nameEn = node.nameEn
    this.localAccuratedName = node.localAccuratedName
    this.character = null
    this.person = null
  }

  public setCharacter(node) {
    this.character = new Character(node)
  }

  public setPerson(node) {
    this.person = new Person(node)
  }
}
