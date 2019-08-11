import Cast from './Cast'
import Episode from './Episode'
import Season from './Season'
import Trailer from './Trailer'
import WorkImage from './WorkImage'

export default class {
  constructor(node) {
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
    this.startedOn = node.startedOn
    this.synopsis = node.synopsis
    this.synopsisEn = node.synopsisEn
    this.synopsisSource = node.synopsisSource
    this.synopsisSourceEn = node.synopsisSourceEn
    this.syobocalTid = node.syobocalTid
    this.title = node.title
    this.titleEn = node.titleEn
    this.titleKana = node.titleKana
    this.twitterHashtag = node.twitterHashtag
    this.twitterUsername = node.twitterUsername
    this.watchersCount = node.watchersCount
    this.wikipediaUrl = node.wikipediaUrl
    this.wikipediaUrlEn = node.wikipediaUrlEn
    this.season = {}
    this.image = {}
    this.casts = []
    this.episodes = []
    this.trailers = []
    this.i18n = null
  }

  setVue(vue) {
    this.vue = vue
  }

  setSeason(node) {
    this.season = new Season(node)
  }

  setImage(node) {
    this.image = new WorkImage(node)
  }

  setCasts(nodes) {
    this.casts = nodes.map(node => {
      const cast = new Cast(node)
      cast.setCharacter(node.character)
      cast.setPerson(node.person)
      return cast
    })
  }

  setEpisodes(nodes) {
    this.episodes = nodes.map(node => {
      return new Episode(node)
    })
  }

  setTrailers(nodes) {
    this.trailers = nodes.map(node => {
      return new Trailer(node)
    })
  }

  localSeasonName() {
    if (this.season.isLater()) {
      return this.vue.$t('models.season.later')
    }

    const seasonName = this.season.name || 'all'

    return this.vue.$t(`models.season.yearly.${seasonName.toLowerCase()}`, { year: this.season.year })
  }

  localStartedOnLabel() {
    if (this.media === 'TV') {
      return this.vue.$t('noun.startToBroadcastTvDate')
    } else if (this.media === 'OVA') {
      return this.vue.$t('noun.startToSellDate')
    } else if (this.media === 'MOVIE') {
      return this.vue.$t('noun.startToBroadcastMovieDate')
    } else {
      return this.vue.$t('noun.startToPublishDate')
    }
  }

  localSynopsis() {
    if (this.vue.$i18n.locale === 'en') {
      return this.synopsisEn
    }

    return this.synopsis
  }
}
