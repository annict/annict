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
                  {{ $t('noun.watchersCount') }}
                </div>
              </div>
              <div class="col text-center">
                <div class="h4 font-weight-bold mb-1">
                  {{ work.satisfactionRate }}<span class="small ml-1">%</span>
                </div>
                <div class="text-muted small">
                  {{ $t('noun.satisfactionRateShorten') }}
                </div>
              </div>
              <div class="col text-center">
                <div class="h4 font-weight-bold mb-1">
                  {{ work.ratingsCount }}
                </div>
                <div class="text-muted small">
                  {{ $t('noun.ratingsCount') }}
                </div>
              </div>
            </div>

            <div class="mb-3">
              <%= render "v3/application/status_selector", gql_work_id: @work.id %>
            </div>

            <h2 class="h4 font-weight-bold mb-3">
              <i class="far fa-sticky-note mr-1"></i>
              {{ $t('noun.information') }}
            </h2>
            <dl>
              <template v-if="work.titleKana">
                <dt class="small">
                  {{ $t('models.work.titleKana') }}
                </dt>
                <dd>
                  {{ work.titleKana }}
                </dd>
              </template>
              <template v-if="work.titleEn">
                <dt class="small">
                  {{ $t('models.work.titleEn') }}
                </dt>
                <dd>
                  {{ work.titleEn }}
                </dd>
              </template>
              <dt class="small">
                {{ $t('models.work.media') }}
              </dt>
              <dd>
                {{ work.media }}
              </dd>
              <template v-if="work.season.year">
                <dt class="small">
                  {{ $t('noun.releaseSeason') }}
                </dt>
                <dd>
                  <a :href="'/works/' + work.season.slug">
                    {{ work.localSeasonName() }}
                  </a>
                </dd>
              </template>
              <template v-if="work.startedOn">
                <dt class="small">
                  {{ work.localStartedOnLabel() }}
                </dt>
                <dd>
                  {{ work.startedOn }}
                </dd>
              </template>
              <template v-if="work.officialSiteUrl">
                <dt class="small">
                  {{ $t('models.work.officialSiteUrl') }}
                </dt>
                <dd>
                  <a :href="work.officialSiteUrl" target="_blank" rel="noopener">
                    {{ work.officialSiteUrl | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.officialSiteUrlEn">
                <dt class="small">
                  {{ $t('models.work.officialSiteUrlEn') }}
                </dt>
                <dd>
                  <a :href="work.officialSiteUrlEn" target="_blank" rel="noopener">
                    {{ work.officialSiteUrlEn | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.twitterUsername">
                <dt class="small">
                  {{ $t('models.work.twitterUsername') }}
                </dt>
                <dd>
                  <a :href="'https://twitter.com/' + work.twitterUsername" target="_blank" rel="noopener">
                    @{{ work.twitterUsername }}
                  </a>
                </dd>
              </template>
              <template v-if="work.twitterHashtag">
                <dt class="small">
                  {{ $t('models.work.twitterHashtag') }}
                </dt>
                <dd>
                  <a :href="'https://twitter.com/search?q=%23' + work.twitterHashtag + '&src=typed_query'" target="_blank" rel="noopener">
                    #{{ work.twitterHashtag }}
                  </a>
                </dd>
              </template>
              <template v-if="work.wikipediaUrl">
                <dt class="small">
                  {{ $t('models.work.wikipediaUrl') }}
                </dt>
                <dd>
                  <a :href="work.wikipediaUrl" target="_blank" rel="noopener">
                    {{ work.wikipediaUrl | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.wikipediaUrlEn">
                <dt class="small">
                  {{ $t('models.work.wikipediaUrlEn') }}
                </dt>
                <dd>
                  <a :href="work.wikipediaUrlEn" target="_blank" rel="noopener">
                    {{ work.wikipediaUrlEn | formatDomain }}
                  </a>
                </dd>
              </template>
              <template v-if="work.syobocalTid">
                <dt class="small">
                  {{ $t('noun.syoboiCalendar') }}
                </dt>
                <dd>
                  <a :href="'http://cal.syoboi.jp/tid/' + work.syobocalTid" target="_blank" rel="noopener">
                    {{ work.syobocalTid }}
                  </a>
                </dd>
              </template>
              <template v-if="work.malAnimeId">
                <dt class="small">
                  {{ $t('noun.myAnimeList') }}
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
                {{ $t('noun.pv') }}
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
                {{ $t('models.work.synopsis') }}
              </h2>
              <div class="c-card mt-3 p-3">
                <span v-html="format(work.localSynopsis())"></span>
                <div class="text-right small">
                  <span class="mr-1">
                    <%= t("noun.source") %>: <%= @work.local_synopsis_source %>
                  </span>
                </div>
              </div>
            </template>

            <h2 class="h4 text-center my-4 font-weight-bold">
              <%= t "noun.episodes" %>
            </h2>
            <div class="c-card container mt-3 pt-3">
              <div class="row">
                <% @work.episodes.each do |episode| %>
                <div class="col-4 mb-3">
                  <%= link_to "" do %>
                  <div class="">
                    <%= episode.number_text %>
                  </div>
                  <div class="small u-text-body">
                    <%= episode.title %>
                  </div>
                  <% end %>
                </div>
                <% end %>
              </div>
            </div>

            <h2 class="h4 text-center my-4 font-weight-bold">
              <%= t "noun.characters" %>
            </h2>
            <div class="c-card container mt-3 pt-3">
              <div class="row">
                <% @work.casts.each do |cast| %>
                <div class="col-3 mb-3">
                  <%= link_to cast.character.name, character_path(cast.character.annict_id) %>
                  <div class="small">
                    <span>CV:</span>
                    <%= cast.decorate.local_accurated_name_link %>
                  </div>
                </div>
                <% end %>
              </div>
            </div>

            <h2 class="h4 text-center my-4 font-weight-bold">
              <%= t "noun.staffs" %>
            </h2>
            <div class="c-card container mt-3 pt-3">
              <div class="row">
                <% @work.staffs.each do |staff| %>
                <div class="col-3 mb-3">
                  <%= staff.decorate.local_accurated_name_link %>
                  <div class="small">
                    <%= staff.local_role %>
                  </div>
                </div>
                <% end %>
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
  import NavBar from '../NavBar.vue'
  import WorkSubNav from '../WorkSubNav.vue'

  import newLine from '../../filters/newLine'

  import { fetchWorkQuery } from '../../queries'

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

    data() {
      return {
        work: null
      }
    },

    methods: {
      format(str) {
        return newLine(str)
      }
    },

    async created() {
      this.work = await fetchWorkQuery({ workId: this.workId })
      this.work.setVue(this)
      console.log('this.work: ', this.work)
    }
  }
</script>
