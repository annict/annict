export default class {
  constructor(node) {
    this.slug = node.seasonSlug
    this.name = node.seasonName
    this.year = node.seasonYear
  }

  isLater() {
    return !this.year && !this.name
  }
}
