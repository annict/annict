import $ from 'jquery'
import 'bootstrap'
import 'select2'
import 'd3'
import 'dropzone'
import {} from 'jquery-ujs'
import Cookies from 'js-cookie'
import moment from 'moment-timezone'
import 'moment/locale/ja'
import Turbolinks from 'turbolinks'
import Vue from 'vue'
import VueLazyload from 'vue-lazyload'

import app from './common/app'
import eventHub from './common/eventHub'
import vueLazyLoad from './common/vueLazyLoad'

import activities from './common/components/activities'
import adsense from './common/components/adsense'
import amazonItemAttacher from './common/components/amazonItemAttacher'
import analytics from './common/components/analytics'
import body from './common/components/body'
import channelReceiveButton from './common/components/channelReceiveButton'
import channelSelector from './common/components/channelSelector'
import commentGuard from './common/components/commentGuard'
import episodeList from './common/components/episodeList'
import episodeProgress from './common/components/episodeProgress'
import episodeRatingStateChart from './common/components/episodeRatingStateChart'
import episodeRecordsChart from './common/components/episodeRecordsChart'
import favoriteButton from './common/components/favoriteButton'
import flash from './common/components/flash'
import followButton from './common/components/followButton'
import impressionButton from './common/components/impressionButton'
import impressionButtonModal from './common/components/impressionButtonModal'
import inputWordsCount from './common/components/inputWordsCount'
import likeButton from './common/components/likeButton'
import omittedSynopsis from './common/components/omittedSynopsis'
import muteUserButton from './common/components/muteUserButton'
import programList from './common/components/programList'
import ratingLabel from './common/components/ratingLabel'
import ratingStateLabel from './common/components/ratingStateLabel'
import reactionButton from './common/components/reactionButton'
import record from './common/components/record'
import recordRating from './common/components/recordRating'
import recordSorter from './common/components/recordSorter'
import recordTextarea from './common/components/recordTextarea'
import recordWordCount from './common/components/recordWordCount'
import shareButtonFacebook from './common/components/shareButtonFacebook'
import shareButtonTwitter from './common/components/shareButtonTwitter'
import statusSelector from './common/components/statusSelector'
import timeAgo from './common/components/timeAgo'
import tips from './common/components/tips'
import untrackedEpisodeList from './common/components/untrackedEpisodeList'
import userHeatmap from './common/components/userHeatmap'
import usernamePreview from './common/components/usernamePreview'
import workComment from './common/components/workComment'
import workFriends from './common/components/workFriends'
import workStatusChart from './common/components/workStatusChart'
import workTags from './common/components/workTags'
import workWatchersChart from './common/components/workWatchersChart'
import youtubeModalPlayer from './common/components/youtubeModalPlayer'

import searchForm from './pc/components/searchForm'
import imageAttachForm from './pc/components/imageAttachForm'
import imageAttachModal from './pc/components/imageAttachModal'

import resourceSelect from './common/directives/resourceSelect'

document.addEventListener('turbolinks:load', event => {
  const gon = window.gon

  moment.locale(gon.user.locale)
  Cookies.set('ann_time_zone', moment.tz.guess(), {
    domain: `.${gon.annict.domain}`,
    secure: gon.rails.env === 'production',
  })

  $.ajaxSetup({
    headers: { 'X-CSRF-Token': $('meta[name="csrf-token"]').attr('content') },
  })

  Vue.config.debug = gon.rails.env !== 'production'

  Vue.use(VueLazyload)

  Vue.nextTick(() => {
    vueLazyLoad.refresh()
  })

  new Vue({
    el: '.p-application',
    components: {
      'c-activities': activities,
      'c-adsense': adsense,
      'c-amazon-item-attacher': amazonItemAttacher,
      'c-analytics': analytics(event),
      'c-body': body,
      'c-channel-receive-button': channelReceiveButton,
      'c-channel-selector': channelSelector,
      'c-comment-guard': commentGuard,
      'c-episode-list': episodeList,
      'c-episode-progress': episodeProgress,
      'c-episode-rating-state-chart': episodeRatingStateChart,
      'c-episode-records-chart': episodeRecordsChart,
      'c-favorite-button': favoriteButton,
      'c-flash': flash,
      'c-follow-button': followButton,
      'c-image-attach-form': imageAttachForm,
      'c-image-attach-modal': imageAttachModal,
      'c-impression-button': impressionButton,
      'c-impression-button-modal': impressionButtonModal,
      'c-input-words-count': inputWordsCount,
      'c-like-button': likeButton,
      'c-omitted-synopsis': omittedSynopsis,
      'c-mute-user-button': muteUserButton,
      'c-program-list': programList,
      'c-rating-label': ratingLabel,
      'c-rating-state-label': ratingStateLabel,
      'c-reaction-button': reactionButton,
      'c-record': record,
      'c-record-rating': recordRating,
      'c-record-sorter': recordSorter,
      'c-record-textarea': recordTextarea,
      'c-record-word-count': recordWordCount,
      'c-search-form': searchForm,
      'c-share-button-facebook': shareButtonFacebook,
      'c-share-button-twitter': shareButtonTwitter,
      'c-status-selector': statusSelector,
      'c-time-ago': timeAgo,
      'c-tips': tips,
      'c-untracked-episode-list': untrackedEpisodeList,
      'c-user-heatmap': userHeatmap,
      'c-username-preview': usernamePreview,
      'c-work-comment': workComment,
      'c-work-friends': workFriends,
      'c-work-status-chart': workStatusChart,
      'c-work-tags': workTags,
      'c-work-watchers-chart': workWatchersChart,
      'c-youtube-modal-player': youtubeModalPlayer,
    },
    directives: {
      'resource-select': resourceSelect,
    },
    data: {
      appData: {},
    },
    created: async function() {
      this.appData = await app.loadAppData()

      if (this.appData.isUserSignedIn) {
        this.pageData = await app.loadPageData()
      }

      eventHub.$emit('app:loaded')
    },
  })
})

Turbolinks.start()
