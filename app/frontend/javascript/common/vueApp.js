import Vue from 'vue';

import favoriteButton from './components/favoriteButton';
import followButton from './components/followButton';
import impressionButton from './components/impressionButton';
import impressionButtonModal from './components/impressionButtonModal';
import userHeatmap from './components/userHeatmap';
import workComment from './components/workComment';
import workFriends from './components/workFriends';
import workTags from './components/workTags';
import youtubeModalPlayer from './components/youtubeModalPlayer';

export default {
  start() {
    Vue.component('c-favorite-button', favoriteButton);
    Vue.component('c-follow-button', followButton);
    Vue.component('c-impression-button', impressionButton);
    Vue.component('c-impression-button-modal', impressionButtonModal);
    Vue.component('c-user-heatmap', userHeatmap);
    Vue.component('c-work-comment', workComment);
    Vue.component('c-work-friends', workFriends);
    Vue.component('c-work-tags', workTags);
    Vue.component('c-youtube-modal-player', youtubeModalPlayer);

    new Vue({
      el: '.ann-application',
    });
  },
};
