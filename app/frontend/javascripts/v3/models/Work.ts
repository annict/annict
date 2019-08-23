import { ApplicationModel, Cast, Episode, Season, Staff, Trailer, WorkImage } from '../models'

export class Work extends ApplicationModel {
  private annictId?: number

  public constructor(node) {
    super()
    this.annictId = node.annictId
    this.copyright = node.copyright
    this.id = node.id
    this.isNoEpisodes = node.isNoEpisodes
    this.malAnimeId = node.malAnimeId
    this.media = node.media
    this.officialSiteUrl = node.officialSiteUrl
    this.officialSiteUrlEn = node.officialSiteUrlEn
    this.ratingsCount = node.ratingsCount
    this.satisfactionRate = node.satisfactionRate
    this.localStartedOnLabel = node.localStartedOnLabel
    this.startedOn = node.startedOn
    this.synopsis = node.synopsis
    this.synopsisEn = node.synopsisEn
    this.localSynopsis = node.localSynopsis
    this.synopsisSource = node.synopsisSource
    this.synopsisSourceEn = node.synopsisSourceEn
    this.localSynopsisSource = node.localSynopsisSource
    this.syobocalTid = node.syobocalTid
    this.title = node.title
    this.titleEn = node.titleEn
    this.localTitle = node.localTitle
    this.titleKana = node.titleKana
    this.twitterHashtag = node.twitterHashtag
    this.twitterUsername = node.twitterUsername
    this.watchersCount = node.watchersCount
    this.wikipediaUrl = node.wikipediaUrl
    this.wikipediaUrlEn = node.wikipediaUrlEn
    this.season = {}
    this.image = {}
    this.casts = []
    this.staffs = []
    this.episodes = []
    this.trailers = []
  }

  public setSeason(node) {
    this.season = new Season(node)
  }

  public setImage(node) {
    this.image = new WorkImage(node)
  }

  public setCasts(nodes) {
    this.casts = nodes.map(node => {
      const cast = new Cast(node)
      cast.setCharacter(node.character)
      cast.setPerson(node.person)
      return cast
    })
  }

  public setStaffs(nodes) {
    this.staffs = nodes.map(node => {
      console.log('node.resource.__typename: ', node.resource.__typename)
      const staff = new Staff(node)
      if (node.resource.__typename === 'Person') {
        staff.setPerson(node.resource)
      } else if (node.resource.__typename === 'Organization') {
        staff.setOrganization(node.resource)
      }
      return staff
    })
  }

  public setEpisodes(nodes) {
    this.episodes = nodes.map(node => {
      return new Episode(node)
    })
  }

  public setTrailers(nodes) {
    this.trailers = nodes.map(node => {
      return new Trailer(node)
    })
  }
}
