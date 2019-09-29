<template>
  <div class="p-footer">
    <div class="p-footer__main pt-3">
      <div class="container">
        <div class="row">
          <div class="col-12 col-sm-3 mb-3">
            <h2 class="mb-0">
              <a :href="AnnConfig.localUrl">Annict</a>
            </h2>
            <div class="p-footer__description small">
              The platform for anime addicts.
            </div>
            <div class="p-footer__social-links">
              <h4></h4>
              <div class="row">
                <div class="col col-auto">
                  <a class="h2" :href="'https://twitter.com/' + AnnConfig.twitter.username" rel="noopener" target="_blank">
                    <i class="fab fa-twitter"></i>
                  </a>
                </div>
                <div class="col col-auto">
                  <a class="h2" href="https://github.com/annict" rel="noopener" target="_blank">
                    <i class="fab fa-github"></i>
                  </a>
                </div>
                <div class="col col-auto">
                  <a class="h2" href="https://discord.gg/Uqbeq8t" rel="noopener" target="_blank">
                    <i class="fab fa-discord"></i></a>
                </div>
              </div>
            </div>
          </div>

          <div class="col-12 col-sm-3">
            <h4 class="font-weight-bold">
              {{ $root.$t('noun.services') }}
            </h4>
            <ul class="p-footer__list list-unstyled">
              <li>
                <a href="/userland">Annict Userland</a>
              </li>
              <li>
                <a href="/forum">Annict Forum</a>
              </li>
              <li>
                <a href="/db">Annict DB</a>
              </li>
              <li>
                <a target="_blank" href="https://developers.annict.jp">Annict Developers</a>
              </li>
              <li>
                <a href="/supporters">Annict Supporter</a>
              </li>
            </ul>
          </div>

          <div class="col-12 col-sm-3">
            <h4 class="font-weight-bold">
              {{ $root.$t('noun.contents') }}
            </h4>
            <ul class="p-footer__list list-unstyled">
              <li>
                <a href="/faqs">
                  {{ $root.$t('head.title.faqs.index') }}
                </a>
              </li>
              <li>
                <a href="/about">
                  {{ $root.$t('head.title.pages.about') }}
                </a>
              </li>
              <li>
                <a href="/terms">
                  {{ $root.$t('head.title.pages.terms') }}
                </a>
              </li>
              <li>
                <a href="/privacy">
                  {{ $root.$t('head.title.pages.privacy') }}
                </a>
              </li>
              <li v-if="$root.isLocaleJa()">
                <a href="/legal">
                  {{ $root.$t('head.title.pages.legal') }}
                </a>
              </li>
            </ul>
          </div>

          <div class="col-12 col-sm-3">
            <h4 class="font-weight-bold">
              {{ $root.$t('noun.seasonalAnime') }}
            </h4>
            <ul class="p-footer__list list-unstyled">
              <li>
                <a href="/works/2019-autumn">
                  {{ $root.$t('noun.yearFall', { year: 2019 }) }}
                </a>
              </li>
              <li>
                <a href="/works/2019-summer">
                  {{ $root.$t('noun.yearSummer', { year: 2019 }) }}
                </a>
              </li>
              <li>
                <a href="/works/2019-spring">
                  {{ $root.$t('noun.yearSpring', { year: 2019 }) }}
                </a>
              </li>
              <li>
                <a href="/works/2019-winter">
                  {{ $root.$t('noun.yearWinter', { year: 2019 }) }}
                </a>
              </li>
              <li>
                <a href="/works/2018-autumn">
                  {{ $root.$t('noun.yearFall', { year: 2018 }) }}
                </a>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>
    <div class="p-footer__auxiliary">
      <div class="container py-2">
        <div class="row align-items-center">
          <div class="col"><h4 class="font-weight-bold d-inline-block mr-2 mb-0">
            {{ $root.$t('noun.languages') }}:
          </h4>
            <ul class="list-inline d-inline-block mb-0">
              <li class="list-inline-item">
                <a :href="AnnConfig.currentUrlJa">
                  日本語
                </a>
              </li>
              <li class="list-inline-item">
                <a :href="AnnConfig.currentUrlEn">
                  English
                </a>
              </li>
            </ul>
          </div>
          <div class="col text-center">
            <a href="/">
              <img width="25" alt="Annict" src="https://d17jazaimr5n2x.cloudfront.net/packs/images/logos/color-white-ca89ea104a92bbd5b93167715ec04138.png" class="img-fluid rounded">
            </a>
          </div>
          <div class="col text-right">
            <div class="p-footer__copyright small">
              <i class="far fa-copyright mr-1"></i>2014-2019 Annict
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
  import { createComponent } from '@vue/composition-api'

  export default createComponent({
    setup(_props, _context) {
      return {
        AnnConfig: window.AnnConfig,
      }
    }
  })
</script>
