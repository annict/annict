import $ from 'jquery';
import _ from 'lodash';

import eventHub from '../../common/eventHub';
import vueLazyLoad from '../../common/vueLazyLoad';

export default {
  template: '#t-untracked-episode-list',

  data() {
    return {
      isLoading: true,
      latestStatuses: [],
      user: null,
      gon: {}
    };
  },

  methods: {
    load() {
      this.latestStatuses = _.each(
        this._latestStatusData().latest_statuses,
        this._initLatestStatus
      );
      this.user = this._latestStatusData().user;
      this.isLoading = false;
      return this.$nextTick(() => vueLazyLoad.refresh());
    },

    filterNoNextEpisode(latestStatuses) {
      return latestStatuses.filter(latestStatus => !!latestStatus.next_episode);
    },

    skipEpisode(latestStatus) {
      if (confirm(this.gon.I18n['messages.tracks.skip_episode_confirmation'])) {
        return $.ajax({
          method: 'PATCH',
          url: `/api/internal/latest_statuses/${latestStatus.id}/skip_episode`
        }).done(latestStatus => {
          const index = this._getLatestStatusIndex(latestStatus);
          return this.$set(
            this.latestStatuses,
            index,
            this._initLatestStatus(latestStatus)
          );
        });
      }
    },

    postRecord(latestStatus) {
      if (latestStatus.record.isSaving) {
        return;
      }

      latestStatus.record.isSaving = true;

      return $.ajax({
        method: 'POST',
        url: '/api/internal/records',
        data: {
          record: {
            episode_id: latestStatus.next_episode.id,
            comment: latestStatus.record.comment,
            shared_twitter: this.user.share_record_to_twitter,
            shared_facebook: this.user.share_record_to_facebook,
            rating_state: latestStatus.record.ratingState
          },
          page_category: this.gon.basic.pageCategory
        }
      })
        .done(() => {
          return $.ajax({
            method: 'GET',
            url: `/api/internal/works/${latestStatus.work.id}/latest_status`
          }).done(newLatestStatus => {
            eventHub.$emit('flash:show', this._flashMessage(latestStatus));
            const index = this._getLatestStatusIndex(newLatestStatus);
            return this.$set(
              this.latestStatuses,
              index,
              this._initLatestStatus(newLatestStatus)
            );
          });
        })
        .fail(function(data) {
          latestStatus.record.isSaving = false;
          const msg =
            (data.responseJSON != null
              ? data.responseJSON.message
              : undefined) || 'Error';
          return eventHub.$emit('flash:show', msg, 'alert');
        });
    },

    _initLatestStatus(latestStatus) {
      latestStatus.record = {
        comment: '',
        isSaving: false,
        ratingState: null,
        isEditingComment: false,
        uid: _.uniqueId(),
        wordCount: 0,
        commentRows: 1
      };
      return latestStatus;
    },

    _getLatestStatusIndex(latestStatus) {
      return _.findIndex(
        this.latestStatuses,
        status => status.id === latestStatus.id
      );
    },

    _flashMessage(latestStatus) {
      const episodeLink = `\
<a href='/works/${latestStatus.work.id}/episodes/${latestStatus.next_episode
        .id}'>
  ${this.gon.I18n['messages.tracks.see_records']}
</a>\
`;
      return `${this.gon.I18n['messages.tracks.tracked']} ${episodeLink}`;
    },

    _latestStatusData() {
      if (!this.gon.latestStatusData) {
        return {};
      }
      return JSON.parse(this.gon.latestStatusData);
    }
  },

  mounted() {
    this.gon = window.gon;
    if (this.gon.user.device === 'pc') {
      $(this.$el).css({ height: window.innerHeight * 0.7 });
    }
    return this.load();
  }
};
