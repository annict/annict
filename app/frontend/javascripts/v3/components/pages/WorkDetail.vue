<template>
  <div>
    <template v-if="work">
      <ann-navbar></ann-navbar>
      <div class="container p-3">
        <div class="row">
          <div class="col-md-3 pr-md-0">
            <div class="c-work-image mb-2">
              <a :href="'/works/' + work.annictId">
                <img :src="work.image.internalUrl" class="img-fluid img-thumbnail rounded">
              </a>
              <div class="u-very-small text-right text-muted">
                <i class="far fa-copyright mr-1"></i>
                {{ work.copyright }}
              </div>
            </div>
            <h1 class="h2 font-weight-bold mb-3">
              {{ work.title }}
            </h1>
            <div class="row mb-3">
              <div class="col text-center">
                <div class="h4 font-weight-bold mb-1">
                  {{ work.watchersCount }}
                </div>
                <div class="text-muted small">
                  {{ $root.$t('noun.watchersCount') }}
                </div>
              </div>
              <div class="col text-center">
                <div class="h4 font-weight-bold mb-1">
                  {{ work.satisfactionRate }}<span class="small ml-1">%</span>
                </div>
                <div class="text-muted small">
                  {{ $root.$t('noun.satisfactionRateShorten') }}
                </div>
              </div>
              <div class="col text-center">
                <div class="h4 font-weight-bold mb-1">
                  {{ work.ratingsCount }}
                </div>
                <div class="text-muted small">
                  {{ $root.$t('noun.ratingsCount') }}
                </div>
              </div>
            </div>

            <div class="mb-3">
              <%= render "v3/application/status_selector", gql_work_id: @work.id %>
            </div>

            <h2 class="h4 font-weight-bold mb-3">
              <i class="far fa-sticky-note mr-1"></i>
              {{ $root.$t('noun.information') }}
            </h2>
            <dl>
              <template v-if="work.titleKana">
                <dt class="small">
                  {{ $root.$t('models.work.titleKana') }}
                </dt>
                <dd>
                  {{ work.titleKana }}
                </dd>
              </template>
              <template v-if="work.titleEn">
                <dt class="small">
                  {{ $root.$t('models.work.titleEn') }}
                </dt>
                <dd>
                  {{ work.titleEn }}
                </dd>
              </template>
              <dt class="small">
                {{ $root.$t('models.work.media') }}
              </dt>
              <dd>
                {{ work.media }}
              </dd>
              <template v-if="work.season.year">
                <dt class="small">
                  {{ $root.$t('noun.releaseSeason') }}
                </dt>
                <dd>
                  <a :href="'/works/' + work.season.slug">
                    {{ work.season.localName }}
                  </a>
                </dd>
              </template>
              <template v-if="work.startedOn">
                <dt class="small">
                  {{ work.localStartedOnLabel }}
                </dt>
                <dd>
                  {{ work.startedOn }}
                </dd>
              </template>
              <template v-if="work.officialSiteUrl">
                <dt class="small">
                  {{ $root.$t('models.work.officialSiteUrl') }}
                </dt>
                <dd>
                  <a :href="work.officialSiteUrl" target="_blank" rel="noopener">
                    {{ work.officialSiteUrl | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.officialSiteUrlEn">
                <dt class="small">
                  {{ $root.$t('models.work.officialSiteUrlEn') }}
                </dt>
                <dd>
                  <a :href="work.officialSiteUrlEn" target="_blank" rel="noopener">
                    {{ work.officialSiteUrlEn | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.twitterUsername">
                <dt class="small">
                  {{ $root.$t('models.work.twitterUsername') }}
                </dt>
                <dd>
                  <a :href="'https://twitter.com/' + work.twitterUsername" target="_blank" rel="noopener">
                    @{{ work.twitterUsername }}
                  </a>
                </dd>
              </template>
              <template v-if="work.twitterHashtag">
                <dt class="small">
                  {{ $root.$t('models.work.twitterHashtag') }}
                </dt>
                <dd>
                  <a :href="'https://twitter.com/search?q=%23' + work.twitterHashtag + '&src=typed_query'" target="_blank" rel="noopener">
                    #{{ work.twitterHashtag }}
                  </a>
                </dd>
              </template>
              <template v-if="work.wikipediaUrl">
                <dt class="small">
                  {{ $root.$t('models.work.wikipediaUrl') }}
                </dt>
                <dd>
                  <a :href="work.wikipediaUrl" target="_blank" rel="noopener">
                    {{ work.wikipediaUrl | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.wikipediaUrlEn">
                <dt class="small">
                  {{ $root.$t('models.work.wikipediaUrlEn') }}
                </dt>
                <dd>
                  <a :href="work.wikipediaUrlEn" target="_blank" rel="noopener">
                    {{ work.wikipediaUrlEn | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.syobocalTid">
                <dt class="small">
                  {{ $root.$t('noun.syoboiCalendar') }}
                </dt>
                <dd>
                  <a :href="'http://cal.syoboi.jp/tid/' + work.syobocalTid" target="_blank" rel="noopener">
                    {{ work.syobocalTid }}
                  </a>
                </dd>
              </template>
              <template v-if="work.malAnimeId">
                <dt class="small">
                  {{ $root.$t('noun.myAnimeList') }}
                </dt>
                <dd>
                  <a :href="'https://myanimelist.net/anime/' + work.malAnimeId" target="_blank" rel="noopener">
                    {{ work.malAnimeId }}
                  </a>
                </dd>
              </template>
            </dl>
          </div>
          <div class="col-md-9">
            <ann-work-subnav :work="work" page-category="workDetail"></ann-work-subnav>

            <template v-if="work.trailers.length">
              <h2 class="h4 text-center my-4 font-weight-bold">
                {{ $root.$t('noun.pv') }}
              </h2>
              <div class="c-card mt-3 pt-3">
                <div class="row ml-3 pr-3">
                  <div class="col-md-4 col-6 text-center mb-3 pl-0" v-for="trailer in work.trailers">
                    <a :href="trailer.url" target="_blank" rel="noopener">
                      <div class="c-video-thumbnail">
                        <div class="c-video-thumbnail__image" :style="'background-image: url(\'' + trailer.internalImageUrl + '\');'"></div>
                        <i class="far fa-play-circle"></i>
                      </div>
                      <div class="small">
                        {{ trailer.title }}
                      </div>
                    </a>
                  </div>
                </div>
              </div>
            </template>

            <template v-if="work.synopsis">
              <h2 class="h4 text-center my-4 font-weight-bold">
                {{ $root.$t('models.work.synopsis') }}
              </h2>
              <div class="c-card mt-3 p-3">
                <span v-html="format(work.localSynopsis)"></span>
                <div class="text-right small">
                  <span class="mr-1">
                    {{ $root.$t('noun.source') }}: {{ work.localSynopsisSource }}
                  </span>
                </div>
              </div>
            </template>

            <h2 class="h4 text-center my-4 font-weight-bold">
              {{ $root.$t('noun.episodes') }}
            </h2>
            <div class="c-card container mt-3 pt-3">
              <div class="row">
                <div class="col-4 mb-3" v-for="episode in work.episodes">
                  <a href="">
                    <div class="">
                      {{ episode.numberText }}
                    </div>
                    <div class="small u-text-body">
                      {{ episode.title }}
                    </div>
                  </a>
                </div>
              </div>
            </div>

            <h2 class="h4 text-center my-4 font-weight-bold">
              {{ $root.$t('noun.characters') }}
            </h2>
            <div class="c-card container mt-3 pt-3">
              <div class="row">
                <div class="col-3 mb-3" v-for="cast in work.casts">
                  <a :href="'/characters/' + cast.character.annictId">
                    {{ cast.character.name }}
                  </a>
                  <div class="small">
                    <span>CV:</span>
                    <a :href="'/people/' + cast.person.annictId">
                      {{ cast.localAccuratedName }}
                    </a>
                  </div>
                </div>
              </div>
            </div>

            <h2 class="h4 text-center my-4 font-weight-bold">
              {{ $root.$t('noun.staffs') }}
            </h2>
            <div class="c-card container mt-3 pt-3">
              <div class="row">
                <div class="col-3 mb-3" v-for="staff in work.staffs">
                  <template v-if="staff.isPerson()">
                    <a :href="'/people/' + staff.person.annictId">
                      {{ staff.localAccuratedName }}
                    </a>
                  </template>
                  <template v-else>
                    <a :href="'/organizations/' + staff.organization.annictId">
                      {{ staff.localAccuratedName }}
                    </a>
                  </template>
                  <div class="small">
                    {{ staff.localRole }}
                  </div>
                </div>
              </div>
            </div>

            <template v-if="$root.$i18n.locale === 'ja'">
              <h2 class="h4 text-center my-4 font-weight-bold">
                {{ $root.$t('noun.vods') }}
              </h2>
              <div class="c-card container mt-3 pt-3">
                <div class="row pr-3">
                  <div class="col-4 mb-3 pr-0" v-for="vodChannel in vodChannels">
                    <template v-if="vodChannel.programs.length > 1">
                      <div class="btn-group w-100">
                        <button class="btn u-btn-link w-100 dropdown-toggle" type="button" data-toggle="dropdown">
                          {{ vodChannel.name }}
                          <div class="dropdown-menu w-100">
                            <template v-for="program in vodChannel.programs">
                              <a :href="program.vodTitleUrl" class="dropdown-item" target="_blank" rel="noopener">
                                {{ program.vodTitleName }}
                              </a>
                            </template>
                          </div>
                        </button>
                      </div>
                    </template>
                    <template v-else-if="vodChannel.programs.length === 1 && vodChannel.programs[0].vodTitleCode">
                      <a :href="vodChannel.programs[0].vodTitleUrl" class="btn u-btn-link w-100" target="_blank" rel="noopener">
                        {{ vodChannel.name }}
                      </a>
                    </template>
                    <template v-else>
                      <button class="btn u-btn-link w-100" type="button" disabled>
                        {{ vodChannel.name }}
                      </button>
                    </template>
                    </div>
                </div>
              </div>
            </template>

            <div class="row align-items-center">
              <div class="col"></div>
              <div class="col">
                <h2 class="h4 text-center my-4 font-weight-bold">
                  {{ $root.$t('noun.recordBodyList') }}
                </h2>
              </div>
              <div class="col text-right">
                <a :href="'/works/' + work.annictId + '/records'" class="btn btn-primary btn-sm">
                  <i class="far fa-edit mr-1"></i>
                  {{ $root.$t('verb.track') }}
                </a>
              </div>
            </div>
            <div class="c-card">
              <template v-if="work.workRecords.length">
                <div class="px-3">
                  <div class="py-3 u-underline" v-for="workRecord in work.workRecords.slice(0, 10)">
                    <div class="row">
                      <div class="col-auto pl-3 pr-0">
                        <a :href="'/@' + workRecord.user.username">
                          <img :src="workRecord.user.avatarUrl" class="img-fluid img-thumbnail rounded-circle">
                        </a>
                      </div>
                      <div class="col">
                        <div class="mb-3">
                          <div class="text-left">
                            <a :href="'/@/' + workRecord.user.username">
                              {{ workRecord.user.name }}
                            </a>
                            <span class="badge.u-badge-supporter.ml-1" v-if="workRecord.user.isSupportoer">
                              {{ $root.$t('noun.supporter') }}
                            </span>
                          </div>
                          <div class="text-left">
                            <a :href="'/@/' + workRecord.user.username + '/records/' + workRecord.record.annictId" class="small text-muted">
                              {{ workRecord.createdAt }}
                            </a>
                            <small class="ml-2 text-muted" v-if="workRecord.modifiedAt">
                              <i class="far pencil-alt"></i>
                            </small>
                            <span class="small ml-2 text-muted">
                              {{ workRecord.record.pageViewsCount }} views
                            </span>
                          </div>
                        </div>
                        <div :class="{ 'p-work-records-show__content clearfix': true, 'c-comment-guard': !work.viewerFinishedToWatch }" @click="removeCommentGuard">
                          <div class="float-right ml-4 mb-4 p-3" v-if="workRecord.ratingOverallState">
                            <div class="small font-weight-bold text-center mb-2">
                              {{ $root.$t('noun.rating') }}
                            </div>
                            <table>
                              <tbody>
                              <tr v-if="workRecord.ratingAnimationState">
                                <th class="font-weight-normal">
                                  {{ $root.$t('noun.animation') }}
                                </th>
                                <td>
                                  <ann-rating-label :init-state="workRecord.ratingAnimationState"></ann-rating-label>
                                </td>
                              </tr>
                              <tr v-if="workRecord.ratingMusicState">
                                <th class="font-weight-normal">
                                  {{ $root.$t('noun.music') }}
                                </th>
                                <td>
                                  <ann-rating-label :init-state="workRecord.ratingMusicState"></ann-rating-label>
                                </td>
                              </tr>
                              <tr v-if="workRecord.ratingStoryState">
                                <th class="font-weight-normal">
                                  {{ $root.$t('noun.story') }}
                                </th>
                                <td>
                                  <ann-rating-label :init-state="workRecord.ratingStoryState"></ann-rating-label>
                                </td>
                              </tr>
                              <tr v-if="workRecord.ratingCharacterState">
                                <th class="font-weight-normal">
                                  {{ $root.$t('noun.character') }}
                                </th>
                                <td>
                                  <ann-rating-label :init-state="workRecord.ratingCharacterState"></ann-rating-label>
                                </td>
                              </tr>
                              <tr v-if="workRecord.ratingOverallState">
                                <th class="font-weight-normal">
                                  {{ $root.$t('noun.overall') }}
                                </th>
                                <td>
                                  <ann-rating-label :init-state="workRecord.ratingOverallState"></ann-rating-label>
                                </td>
                              </tr>
                              </tbody>
                            </table>
                          </div>
                          <div class="c-body mb-3">
                            <div class="c-body__content" v-html="format(workRecord.body)"></div>
                          </div>
                        </div>

                        <div class="row align-items-center">
                          <div class="col">
                            <div class="text-right">
                              <span class="mr-2">
                                <ann-share-to-twitter-button :text="$root.$t('head.title.workRecords.show', { profileName: workRecord.user.name, username: workRecord.user.username, workTitle: work.localTitle })" :url="annConfig.localUrl + '/@' + workRecord.user.username + '/records/' + workRecord.record.annictId" :hashtags="work.twitterHashtag || ''"></ann-share-to-twitter-button>
                              </span>
                              <span class="mr-2">
                                <ann-share-to-facebook-button :url="annConfig.localUrl + '/@' + workRecord.user.username + '/records/' + workRecord.record.annictId"></ann-share-to-facebook-button>
                              </span>
                            </div>
                          </div>
                        </div>

                        <div class="small text-right mt-2" v-if="$root.viewer.username === workRecord.user.username">
                          <a :href="'/works/' + work.annictId + '/records/' + workRecord.record.annictId + '/edit'" class="mr-2">
                            <i class="fab fa-edit mr-1"></i>
                            {{ $root.$t('noun.edit') }}
                          </a>
                          <a :href="'/@' + workRecord.user.username + '/records/' + workRecord.record.annictId">
                            <i class="far fa-trash-alt mr-1"></i>
                            {{ $root.$t('noun.delete') }}
                          </a>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="container mb-3" v-if="work.workRecords.length > 10">
                    <a :href="'/works/' + work.annictId + '/records'" class="btn btn-secondary w-100">
                      <i class="fab fa-angle-right"></i>
                      {{ $root.$t('messages.works.viewAllNRecordBodyList', { n: work.workRecords.length }) }}
                    </a>
                  </div>
                </div>
              </template>
              <template v-else>
                <ann-empty :text="$root.$t('messages._components.empty.noRecordBodyList')">
                  <a :href="'/works/' + work.annictId + '/records'" class="btn btn-primary mt-2">
                    <i class="fab fa-edit mr-1"></i>
                    {{ $root.$t('verb.track') }}
                  </a>
                </ann-empty>
              </template>
            </div>
          </div>
        </div>
      </div>
    </template>
    <template v-else>
      Loading...
    </template>
  </div>
</template>

<script lang="ts">
  import $ from 'jquery'
  import { onCreated, value } from 'vue-function-api'

  import Empty from '../Empty.vue'
  import NavBar from '../NavBar.vue'
  import RatingLabel from '../RatingLabel.vue'
  import ShareToFacebookButton from '../ShareToFacebookButton.vue'
  import ShareToTwitterButton from '../ShareToTwitterButton.vue'
  import WorkSubNav from '../WorkSubNav.vue'

  import escape from '../../filters/escape'
  import newLine from '../../filters/newLine'

  import { FetchWorkQuery } from '../../queries'
  import { FetchVodChannelsQuery } from '../../queries/fetchVodChannelsQuery'

  export default {
    components: {
      'ann-empty': Empty,
      'ann-navbar': NavBar,
      'ann-rating-label': RatingLabel,
      'ann-share-to-facebook-button': ShareToFacebookButton,
      'ann-share-to-twitter-button': ShareToTwitterButton,
      'ann-work-subnav': WorkSubNav,
    },

    props: {
      workId: {
        type: Number,
        required: true
      }
    },

    setup(props, _context) {
      const workWrapper = value(null)
      const vodChannelsWrapper = value([])

      const format = (str) => {
        return newLine(escape(str))
      }

      const removeCommentGuard = (event) => {
        $(event.target).parents('.p-work-records-show__content').removeClass('c-comment-guard')
      }

      onCreated(async () => {
        const [work, vodChannels] = await Promise.all([
          new FetchWorkQuery({ workId: props.workId }).execute(),
          new FetchVodChannelsQuery().execute()
        ])
        console.log('work: ', work)
        workWrapper.value = work
        vodChannelsWrapper.value = vodChannels.map(vodChannel => {
          vodChannel.setProgramsOfWork(work)
          return vodChannel
        })
      })

      return {
        annConfig: window.annConfig,
        work: workWrapper,
        vodChannels: vodChannelsWrapper,
        format,
        removeCommentGuard,
      }
    }
  }
</script>
