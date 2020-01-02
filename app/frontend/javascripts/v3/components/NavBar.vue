<template>
  <div class="c-navbar">
    <!--  PC-->
    <nav class="navbar navbar-expand navbar-light bg-white d-sm-flex d-none">
      <a href="/" class="navbar-brand">
        <img :src="state.AnnConfig.images.logoUrl" width="25" height="30" alt="Annict">
      </a>
      <ul class="navbar-nav mt-2 mt-md-0 mr-md-2">
        <li class="nav-item" v-if="$root.isSignedIn()">
          <a class="nav-link text-dark" href="/programs">
            {{ $t('noun.slots') }}
          </a>
        </li>
        <li class="nav-item dropdown" v-if="$root.isSignedIn()">
          <a class="nav-link dropdown-toggle text-dark" data-toggle="dropdown" href="">
            {{ $t('noun.library') }}
          </a>
          <div class="dropdown-menu m-0">
            <a :href="'/@' + $root.viewer.username + '/watching'" class="dropdown-item text-dark">
              {{ $t('noun.watching') }}
            </a>
            <a :href="'/@' + $root.viewer.username + '/wanna_watch'" class="dropdown-item text-dark">
              {{ $t('noun.planToWatch') }}
            </a>
            <a :href="'/@' + $root.viewer.username + '/watched'" class="dropdown-item text-dark">
              {{ $t('noun.completed') }}
            </a>
            <a :href="'/@' + $root.viewer.username + '/on_hold'" class="dropdown-item text-dark">
              {{ $t('noun.onHold') }}
            </a>
            <a :href="'/@' + $root.viewer.username + '/stop_watching'" class="dropdown-item text-dark">
              {{ $t('noun.dropped') }}
            </a>
          </div>
        </li>
        <li class="nav-item dropdown">
          <a class="nav-link dropdown-toggle text-dark" href="" data-toggle="dropdown">
            {{ $t('verb.explore') }}
          </a>
          <div class="dropdown-menu">
            <a :href="'/works/' + state.AnnConfig.season.current" class="dropdown-item">
              {{ $t('noun.currentSeason') }}
            </a>
            <a :href="'/works/' + state.AnnConfig.season.next" class="dropdown-item">
              {{ $t('noun.nextSeason') }}
            </a>
            <a :href="'/works/' + state.AnnConfig.season.prev" class="dropdown-item">
              {{ $t('noun.prevSeason') }}
            </a>
            <a href="/works/newest" class="dropdown-item">
              {{ $t('head.title.works.newest') }}
            </a>
            <a href="/search" class="dropdown-item">
              {{ $t('verb.search') }}
            </a>
          </div>
        </li>
      </ul>
      <form action="/search" autocomplete="off" class="col-2 px-0 mr-auto" method="get">
        <input name="q" class="form-control" type="text" :placeholder="$t('messages._common.searchWithKeywords')">
      </form>
      <ul class="navbar-nav">
        <template v-if="$root.isSignedIn()">
          <a href="/track" class="btn btn-outline-primary">
            <i class="far fa-edit"></i>{{ $t('verb.track') }}
          </a>
          <a href="#" class="nav-link dropdown-toggle p-0 text-dark" data-toggle="dropdown">
            <img :src="$root.viewer.avatarUrl" width="30" height="30" :alt="$root.viewer.username" class="rounded-circle">
          </a>
          <div class="dropdown-menu dropdown-menu-right">
            <a :href="'/@' + $root.viewer.username" class="dropdown-item">
              {{ $t('noun.profile') }}
            </a>
            <a href="/notifications" class="dropdown-item">
              {{ $t('head.title.notifications.index') }}
            </a>
            <a href="/friends" class="dropdown-item">
              {{ $t('head.title.friends.index') }}
            </a>
            <template v-if="$root.viewer.locale === 'ja'">
              <a href="/channels" class="dropdown-item">
                {{ $t('head.title.channels.index') }}
              </a>
            </template>
            <a href="/settings/profile" class="dropdown-item">
              {{ $t('noun.settings') }}
            </a>
            <a href="/userland" class="dropdown-item">
              {{ $t('noun.annictUserland') }}
            </a>
            <a href="/forum" class="dropdown-item">
              {{ $t('noun.annictForum') }}
            </a>
            <a href="/db" class="dropdown-item">
              {{ $t('noun.annictDb') }}
            </a>
            <template v-if="$root.viewer.locale === 'ja'">
              <a href="https://developers.annict.jp" class="dropdown-item" target="_blank">
                {{ $t('noun.annictDevelopers') }}
              </a>
            </template>
            <a href="/supporters" class="dropdown-item">
              {{ $t('noun.annictSupporters') }}
            </a>
            <template v-if="$root.viewer.locale === 'ja'">
              <a href="/faqs" class="dropdown-item">
                {{ $t('head.title.faqs.index') }}
              </a>
            </template>
            <a href="/about" class="dropdown-item">
              {{ $t('head.title.pages.about') }}
            </a>
            <a data-method="delete" href="/sign_out" class="dropdown-item">
              {{ $t('verb.signOut') }}
            </a>
          </div>
        </template>
        <template v-else>
          <li class="nav-item">
            <a href="/about" class="nav-link text-dark">
              {{ $t('head.title.pages.about') }}
            </a>
          </li>
          <li class="nav-item">
            <a href="/sign_in" class="nav-link text-dark">
              {{ $t('noun.signIn') }}
            </a>
          </li>
          <li class="nav-item">
            <a href="/sign_up" class="btn btn btn-outline-primary">
              <i class="fas fa-rocket"></i>{{ $t('noun.signUp') }}
            </a>
          </li>
        </template>
      </ul>
    </nav>

    <!--  Mobile-->
    <nav class="bg-white d-block d-sm-none h-100 navbar navbar-expand navbar-white px-0">
      <ul class="navbar-nav justify-content-around align-items-center h-100">
        <li class="nav-item text-center col px-0">
          <a href="/" class="text-dark">
            <i class="fas fa-home"></i>
            <div class="small mt-1">
              {{ $t('noun.home') }}
            </div>
          </a>
        </li>
        <template v-if="$root.isSignedIn()">
          <li class="nav-item text-center col px-0">
            <a href="/programs" class="text-dark">
              <i class="far fa-calendar"></i>
              <div class="small mt-1">
                {{ $t('noun.slots') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center col px-0">
            <a class="text-dark" :href="'/@' + $root.viewer.username + '/watching'">
              <i class="fas fa-play"></i>
              <div class="small mt-1">
                {{ $t('noun.watchingShorten') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center col px-0">
            <a :href="'/works/' + state.AnnConfig.season.current" class="text-dark">
              <i class="fas fa-tv"></i>
              <div class="small mt-1">
                {{ $t('noun.airing') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center col px-0">
            <a href="/menu" class="text-dark">
              <i class="fas fa-th"></i>
              <div class="small mt-1">
                {{ $t('noun.menu') }}
              </div>
            </a>
          </li>
        </template>
        <template v-else>
          <li class="nav-item text-center col px-0">
            <a :href="'/works/' + state.AnnConfig.season.current" class="text-dark">
              <i class="fas fa-tv"></i>
              <div class="small mt-1">
                {{ $t('noun.airing') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center col px-0">
            <a href="/sign_up" class="text-dark">
              <i class="fas fa-rocket"></i>
              <div class="small mt-1">
                {{ $t('noun.signUpShorten') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center col px-0">
            <a href="/about" class="text-dark">
              <i class="far fa-lightbulb"></i>
              <div class="small mt-1">
                {{ $t('noun.about') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center col px-0">
            <a href="/menu" class="text-dark">
              <i class="fas fa-th"></i>
              <div class="small mt-1">
                {{ $t('noun.menu') }}
              </div>
            </a>
          </li>
        </template>
      </ul>
    </nav>
  </div>
</template>

<script lang="ts">
  import { createComponent, reactive } from '@vue/composition-api'

  export default createComponent({
    setup() {
      const state = reactive({
        AnnConfig: window.AnnConfig
      })

      return {
        state
      }
    }
  })
</script>
