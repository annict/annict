import { ApplicationModel, Organization, Person } from '../models'

export class Staff extends ApplicationModel {
  public name: string
  public nameEn: string
  public localAccuratedName: string
  public role: string
  public roleEn: string
  public localRole: string
  public organization: Organization
  public person: Person

  public constructor(node) {
    super()
    this.name = node.name
    this.nameEn = node.nameEn
    this.localAccuratedName = node.localAccuratedName
    this.role = node.role
    this.roleEn = node.roleEn
    this.localRole = node.localRole
    this.organization = null
    this.person = null
  }

  public setOrganization(node) {
    this.organization = new Organization(node)
  }

  public setPerson(node) {
    this.person = new Person(node)
  }

  public isPerson() {
    return !!this.person
  }
}
