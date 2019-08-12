import { ApplicationModel, Organization, Person } from '../models'

export class Staff extends ApplicationModel {
  public constructor(obj) {
    super()
    this.name = obj.name
    this.nameEn = obj.nameEn
    this.role = obj.role
    this.roleEn = obj.roleEn
    this.organization = {}
    this.person = {}
  }

  public setOrganization(node) {
    this.organization = new Organization(node)
  }

  public setPerson(node) {
    this.person = new Person(node)
  }

  public isPerson() {
    return !!this.person.annictId
  }
}
