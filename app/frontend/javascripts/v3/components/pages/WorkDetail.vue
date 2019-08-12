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
                    {{ localSeasonName(work) }}
                  </a>
                </dd>
              </template>
              <template v-if="work.startedOn">
                <dt class="small">
                  {{ localStartedOnLabel(work) }}
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
                <span v-html="format(localField(work, 'synopsis'))"></span>
                <div class="text-right small">
                  <span class="mr-1">
                    {{ $root.$t('noun.source') }}: {{ localField(work, 'synopsisSource') }}
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
                      {{ localAccuratedPersonName(cast) }}
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
                      {{ localAccuratedPersonName(staff) }}
                    </a>
                  </template>
                  <template v-else>
                    <a :href="'/organizations/' + staff.organization.annictId">
                      {{ localAccuratedOrgName(staff) }}
                    </a>
                  </template>
                  <div class="small">
                    {{ localField(staff, 'role') }}
                  </div>
                </div>
              </div>
            </div>

            <% if locale_ja? %>
            <h2 class="h4 text-center my-4 font-weight-bold">
              <%= t "noun.vods" %>
            </h2>
            <div class="c-card container mt-3 pt-3">
              <%= render "v3/works/vod_list", vod_channels: @vod_channels, programs: @work.programs %>
            </div>
            <% end %>

            <div class="row align-items-center">
              <div class="col"></div>
              <div class="col">
                <h2 class="h4 text-center my-4 font-weight-bold"></h2>
                <%= t "noun.record_body_list" %>
              </div>
              <div class="col text-right">
                <%= link_to work_records_path(@work.annict_id), class: "btn btn-primary btn-sm" do %>
                <%= icon "edit", class: "mr-1" %>
                <%= t "verb.track" %>
                <% end %>
              </div>
            </div>
            <div class="c-card">
              <%= render "v3/works/work_record_list", work: @work, work_records: @work_records %>
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
  import { onCreated, value } from 'vue-function-api'

  import NavBar from '../NavBar.vue'
  import WorkSubNav from '../WorkSubNav.vue'

  import newLine from '../../filters/newLine'

  import { FetchWorkQuery } from '../../queries'

  export default {
    components: {
      'ann-navbar': NavBar,
      'ann-work-subnav': WorkSubNav,
    },

    props: {
      workId: {
        type: Number,
        required: true
      }
    },

    setup(props, context) {
      const work = value(null)

      const format = (str) => {
        return newLine(str)
      }

      const localField = (model, fieldName) => {
        if (context.root.$i18n.locale === 'en') {
          return model[`${fieldName}En`]
        }

        return model[fieldName]
      }

      const localSeasonName = (work) => {
        if (work.season.isLater()) {
          return context.root.$t('models.season.later')
        }

        const seasonName = work.season.name || 'all'

        return context.root.$t(`models.season.yearly.${seasonName.toLowerCase()}`, { year: work.season.year })
      }

      const localStartedOnLabel = (work) => {
        if (work.media === 'TV') {
          return context.root.$t('noun.startToBroadcastTvDate')
        } else if (work.media === 'OVA') {
          return context.root.$t('noun.startToSellDate')
        } else if (work.media === 'MOVIE') {
          return context.root.$t('noun.startToBroadcastMovieDate')
        } else {
          return context.root.$t('noun.startToPublishDate')
        }
      }

      const localAccuratedPersonName = (model) => {
        const locale = context.root.$i18n.locale

        if (localField(model, 'name') === localField(model.person, 'name')) {
          return localField(model, 'name')
        }

        `${localField(model, 'name')} (${localField(model.person, 'name')})`
      }

      const localAccuratedOrgName = (model) => {
        const locale = context.root.$i18n.locale

        if (localField(model, 'name') === localField(model.organization, 'name')) {
          return localField(model, 'name')
        }

        `${localField(model, 'name')} (${localField(model.organization, 'name')})`
      }

      onCreated(async () => {
        work.value = await new FetchWorkQuery({ workId: props.workId }).execute()
        console.log('work: ', work)
      })

      return {
        work,
        format,
        localField,
        localSeasonName,
        localStartedOnLabel,
        localAccuratedPersonName,
        localAccuratedOrgName,
      }
    }
  }
</script>
