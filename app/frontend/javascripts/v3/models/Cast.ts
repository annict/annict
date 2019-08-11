import Character from './Character'
import Person from './Person'

export default class {
  constructor(obj) {
    this.name = obj.name
    this.nameEn = obj.nameEn
    this.character = {}
    this.person = {}
  }

  setCharacter(node) {
    this.character = new Character(node)
  }

  setPerson(node) {
    this.person = new Person(node)
  }
}
