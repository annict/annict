import { ApplicationModel, Cast, Episode, Program, Season, Staff, Trailer, WorkImage, WorkRecord } from '../models'

export class Work extends ApplicationModel {
  private annictId?: number
  private copyright: string
  private id: string
  private isNoEpisodes: boolean
  private malAnimeId?: number
  private media: string
  private officialSiteUrl: string
  private officialSiteUrlEn: string
  private ratingsCount: number
  private satisfactionRate: number
  private localStartedOnLabel: string
  private startedOn: date
  private synopsis: string
  private synopsisEn: string
  private localSynopsis: string
  private localSynopsisHtml: string
  private synopsisSource: string
  private synopsisSourceEn: string
  private localSynopsisSource: string
  private syobocalTid: number
  private title: string
  private titleEn: string
  private localTitle: string
  private titleKana: string
  private twitterHashtag: string
  private twitterUsername: string
  private watchersCount: number
  private wikipediaUrl: string
  private wikipediaUrlEn: string
  private viewerFinishedToWatch: boolean
  private viewerStatusKind: string
  private season: Season
  private image: WorkImage
  private trailers: Trailer[]
  private casts: Cast[]
  private staffs: Staff[]
  private episodes: Episode[]
  private programs: Program[]
  private workRecords: WorkRecord[]

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
    this.localSynopsisHtml = node.localSynopsisHtml
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
    this.wikipediaUrl = node.wikipediaUrl
    this.wikipediaUrlEn = node.wikipediaUrlEn
    this.viewerFinishedToWatch = node.viewerFinishedToWatch
    this.viewerStatusKind = node.viewerStatusKind
    this.watchersCount = node.watchersCount
  }

  public setSeason(node) {
    this.season = new Season(node)
  }

  public setImage(node) {
    this.image = new WorkImage(node)
  }

  public setTrailers(nodes) {
    this.trailers = nodes.map(node => {
      return new Trailer(node)
    })
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

  public setPrograms(nodes) {
    this.programs = nodes.map(node => {
      const program = new Program(node)
      program.setChannel(node.channel)
      return program
    })
  }

  public setWorkRecords(nodes) {
    this.workRecords = nodes.map(node => {
      const workRecord = new WorkRecord(node)
      workRecord.setUser(node.user)
      workRecord.setRecord(node.record)
      return workRecord
    })
  }
}
