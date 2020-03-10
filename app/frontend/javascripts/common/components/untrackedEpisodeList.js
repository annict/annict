import $ from 'jquery';
import _ from 'lodash';

import eventHub from '../../common/eventHub';
import vueLazyLoad from '../../common/vueLazyLoad';

export default {
  template: '#t-untracked-episode-list',

  data() {
    return {
      isLoading: true,
      libraryEntries: [],
      user: null,
      gon: {},
    };
  },

  methods: {
    load() {
      this.libraryEntries = _.each(this._latestStatusData().library_entries, this._initLibraryEntry);
      this.user = this._latestStatusData().user;
      this.isLoading = false;
      return this.$nextTick(() => vueLazyLoad.refresh());
    },

    filterNoNextEpisode(libraryEntries) {
      return libraryEntries.filter(latestStatus => !!latestStatus.next_episode);
    },

    skipEpisode(latestStatus) {
      if (confirm(this.gon.I18n['messages.tracks.skip_episode_confirmation'])) {
        return $.ajax({
          method: 'PATCH',
          url: `/api/internal/library_entries/${latestStatus.id}/skip_episode`,
        }).done(latestStatus => {
          const index = this._getLibraryEntryIndex(latestStatus);
          return this.$set(this.libraryEntries, index, this._initLibraryEntry(latestStatus));
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
        url: `/api/internal/episodes/${latestStatus.next_episode.id}/records`,
        data: {
          episode_record: {
            body: latestStatus.record.body,
            shared_twitter: this.user.share_record_to_twitter,
            rating_state: latestStatus.record.ratingState,
          },
          page_category: this.gon.page.category,
        },
      })
        .done(() => {
          return $.ajax({
            method: 'GET',
            url: `/api/internal/works/${latestStatus.work.id}/library_entry`,
          }).done(newLibraryEntry => {
            eventHub.$emit('flash:show', this._flashMessage(latestStatus));
            const index = this._getLibraryEntryIndex(newLibraryEntry);
            return this.$set(this.libraryEntries, index, this._initLibraryEntry(newLibraryEntry));
          });
        })
        .fail(function(data) {
          latestStatus.record.isSaving = false;
          const msg = (data.responseJSON != null ? data.responseJSON.message : undefined) || 'Error';
          return eventHub.$emit('flash:show', msg, 'alert');
        });
    },

    _initLibraryEntry(latestStatus) {
      latestStatus.record = {
        comment: '',
        isSaving: false,
        ratingState: null,
        isEditingBody: false,
        uid: _.uniqueId(),
        wordCount: 0,
        commentRows: 1,
      };
      return latestStatus;
    },

    _getLibraryEntryIndex(latestStatus) {
      return _.findIndex(this.libraryEntries, status => status.id === latestStatus.id);
    },

    _flashMessage(latestStatus) {
      const episodeLink = `\
<a href='/works/${latestStatus.work.id}/episodes/${latestStatus.next_episode.id}'>
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
    },
  },

  mounted() {
    this.gon = window.gon;
    return this.load();
  },
};
