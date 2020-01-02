import { ApplicationModel, Cast, Episode, Program, Season, Series, Staff, Trailer, WorkImage, WorkRecord } from '../models'

export class Work extends ApplicationModel {
  public annictId?: number
  public copyright: string
  public id: string
  public isNoEpisodes: boolean
  public malAnimeId?: number
  public media: string
  public officialSiteUrl: string
  public officialSiteUrlEn: string
  public episodesCount: number
  public ratingsCount: number
  public satisfactionRate: number
  public localStartedOnLabel: string
  public startedOn: string
  public synopsis: string
  public synopsisEn: string
  public localSynopsis: string
  public localSynopsisHtml: string
  public synopsisSource: string
  public synopsisSourceEn: string
  public localSynopsisSource: string
  public syobocalTid: number
  public title: string
  public titleAlter: string
  public titleAlterEn: string
  public titleEn: string
  public localTitle: string
  public titleKana: string
  public twitterHashtag: string
  public twitterUsername: string
  public watchersCount: number
  public workRecordsWithBodyCount: number
  public wikipediaUrl: string
  public wikipediaUrlEn: string
  public viewerFinishedToWatch: boolean
  public viewerStatusKind: string
  public season: Season
  public image: WorkImage
  public trailers: Trailer[]
  public casts: Cast[]
  public staffs: Staff[]
  public episodes: Episode[]
  public programs: Program[]
  public workRecords: WorkRecord[]
  public seriesList: Series[]

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
    this.episodesCount = node.episodesCount
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
    this.titleAlter = node.titleAlter
    this.titleAlterEn = node.titleAlterEn
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
    this.workRecordsWithBodyCount = node.workRecordsWithBodyCount
    this.season = new Season(node)
    this.image = new WorkImage(node.image)
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

  public setSeriesList(nodes) {
    this.seriesList = nodes.map(node => {
      const series = new Series(node)
      series.setSeriesWorks(node.works.edges)
      return series
    })
  }
}
