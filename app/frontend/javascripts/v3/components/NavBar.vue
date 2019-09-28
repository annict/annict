<template>
  <div class="c-navbar">
    <!--  PC-->
    <nav class="navbar navbar-expand navbar-light bg-white">
      <a href="/" class="navbar-brand d-none d-lg-inline-block">
        <img :src="state.annConfig.images.logoUrl" width="25" height="30" alt="Annict">
      </a>
      <ul class="navbar-nav mt-2 mt-md-0 mr-md-2">
        <li class="nav-item" v-if="$root.isSignedIn()">
          <a class="nav-link text-dark" href="/programs">
            {{ $t('noun.programs') }}
          </a>
        </li>
        <li class="nav-item dropdown d-none d-lg-inline-block" v-if="$root.isSignedIn()">
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
        <li class="nav-item dropdown d-none d-lg-inline-block">
          <a class="nav-link dropdown-toggle text-dark" href="" data-toggle="dropdown">
            {{ $t('verb.explore') }}
          </a>
          <div class="dropdown-menu">
            <a :href="'/works/' + state.annConfig.season.current" class="dropdown-item">
              {{ $t('noun.currentSeason') }}
            </a>
            <a :href="'/works/' + state.annConfig.season.next" class="dropdown-item">
              {{ $t('noun.nextSeason') }}
            </a>
            <a :href="'/works/' + state.annConfig.season.prev" class="dropdown-item">
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
      <form action="/search" autocomplete="off" class="col-md-2 px-0 mr-auto d-none d-lg-inline-block" method="get">
        <input name="q" class="form-control" type="text" :placeholder="$t('messages._common.searchWithKeywords')">
      </form>
      <ul class="navbar-nav">
        <template v-if="$root.isSignedIn()">
          <a href="#" class="nav-link dropdown-toggle p-0 text-dark" data-toggle="dropdown">
            <img :src="$root.viewer.avatarUrl" width="30" height="30" :alt="$root.viewer.username" class="rounded-circle">
          </a>
          <div class="dropdown-menu dropdown-menu-right">
            <a :href="'/@' + $root.viewer.username" class="dropdown-item">
              {{ $t('noun.profile') }}
            </a>
            <a href="/friends" class="dropdown-item">
              {{ $t('head.title.friends.index') }}
            </a>
            <template v-if="$root.viewer.locale === 'ja'">
              <a href="/channels" class="dropdown-item">
                {{ $t('head.title.channels.index') }}
              </a>
            </template>
            <a href="/profile" class="dropdown-item">
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
            <a href="#" class="dropdown-item">
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
    <nav class="navbar navbar-expand navbar-white bg-white fixed-bottom d-block d-lg-none">
      <ul class="navbar-nav justify-content-around align-items-center h-100">
        <li class="nav-item text-center">
          <a href="/" class="text-dark">
            <i class="fas fa-home"></i>
            <div class="small mt-1">
              {{ $t('noun.home') }}
            </div>
          </a>
        </li>
        <template v-if="$root.isSignedIn()">
          <li class="nav-item text-center">
            <a href="/programs" class="text-dark">
              <i class="far fa-calendar"></i>
              <div class="small mt-1">
                {{ $t('noun.programs') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center">
            <a class="text-dark" :href="'/@' + $root.viewer.username + '/watching'">
              <i class="fas fa-play"></i>
              <div class="small mt-1">
                {{ $t('noun.watchingShorten') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center">
            <a :href="'/works/' + state.annConfig.season.current" class="text-dark">
              <i class="fas fa-tv"></i>
              <div class="small mt-1">
                {{ $t('noun.airing') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center position-relative">
            <a href="/menu" class="text-dark">
              <i class="fas fa-th"></i>
              <div class="small mt-1">
                {{ $t('noun.menu') }}
              </div>
            </a>
          </li>
        </template>
        <template v-else>
          <li class="nav-item text-center">
            <a :href="'/works/' + state.annConfig.season.current" class="text-dark">
              <i class="fas fa-tv"></i>
              <div class="small mt-1">
                {{ $t('noun.airing') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center">
            <a href="/sign_up" class="text-dark">
              <i class="fas fa-rocket"></i>
              <div class="small mt-1">
                {{ $t('noun.signUpShorten') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center">
            <a href="/about" class="text-dark">
              <i class="far fa-lightbulb"></i>
              <div class="small mt-1">
                {{ $t('noun.about') }}
              </div>
            </a>
          </li>
          <li class="nav-item text-center">
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
        annConfig: window.annConfig
      })

      return {
        state
      }
    }
  })
</script>
