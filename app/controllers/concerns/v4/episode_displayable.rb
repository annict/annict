# frozen_string_literal: true

module V4
  module EpisodeDisplayable
    def load_episode_and_records(work_id:, episode_id:, form: nil)
      anime = Work.only_kept.find(work_id)
      episode = anime.episodes.only_kept.find(episode_id)

      @episode_entity = Rails.cache.fetch(episode_cache_key(episode), expires_in: 3.days) do
        EpisodePage::EpisodeRepository.new(graphql_client: graphql_client).execute(database_id: episode.id)
      end
      @other_record_entities = @episode_entity.records

      load_vod_channel_entities(anime: anime, anime_entity: @episode_entity.anime)

      @form = form.presence || EpisodeRecordForm.new(episode_id: @episode_entity.id)

      if user_signed_in?
        @my_record_entities = Rails.cache.fetch(my_records_cache_key(current_user, episode), expires_in: 1.day) do
          EpisodePage::MyRecordsRepository.new(graphql_client: graphql_client(viewer: current_user)).execute(episode_id: @episode_entity.id)
        end

        @following_record_entities = Rails.cache.fetch(following_records_cache_key(current_user, episode), expires_in: 3.hours) do
          EpisodePage::FollowingRecordsRepository.new(graphql_client: graphql_client(viewer: current_user)).execute(episode_id: @episode_entity.id)
        end

        @other_record_entities = filter_records(current_user, @other_record_entities, @my_record_entities + @following_record_entities)
      end
    end

    private

    def episode_cache_key(episode)
      [
        PageCategory::EPISODE,
        "episode",
        episode.id,
        episode.updated_at.rfc3339,
        episode.last_record_watched_at&.rfc3339
      ].freeze
    end

    def my_records_cache_key(user, episode)
      [
        PageCategory::EPISODE,
        "my-records",
        episode.id,
        user.id,
        user.last_record_watched_at&.rfc3339
      ].freeze
    end

    def following_records_cache_key(user, episode)
      [
        PageCategory::EPISODE,
        "following-records",
        episode.id,
        user.id
      ].freeze
    end

    def filter_records(viewer, base_record_entities, record_entities)
      muted_user_ids = viewer.mute_users.pluck(:muted_user_id)
      record_ids = record_entities.pluck(:database_id)

      base_record_entities.filter do |record_entity|
        user_id = record_entity.user.database_id
        record_id = record_entity.database_id

        !user_id.in?(muted_user_ids) && !record_id.in?(record_ids)
      end
    end
  end
end
