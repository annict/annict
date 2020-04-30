# frozen_string_literal: true

module WorkDetail
  class WorkEntity < ApplicationEntity
    local_attributes :title, :title_alter, :synopsis_source

    attribute :id, Types::Integer
    attribute :title, Types::String
    attribute :title_en, Types::String.optional
    attribute :title_kana, Types::String.optional
    attribute :title_alter, Types::String.optional
    attribute :title_alter_en, Types::String.optional
    attribute :media, Types::WorkMediaKinds
    attribute :season_year, Types::Integer.optional
    attribute :season_type, Types::SeasonKinds.optional
    attribute :season_slug, Types::String.optional
    attribute :started_on, Types::Date.optional
    attribute :episodes_count, Types::Integer
    attribute :watchers_count, Types::Integer
    attribute :satisfaction_rate, Types::Float.optional
    attribute :ratings_count, Types::Integer
    attribute :work_records_with_body_count, Types::Integer
    attribute :official_site_url, Types::String.optional
    attribute :official_site_url_en, Types::String.optional
    attribute :wikipedia_url, Types::String.optional
    attribute :wikipedia_url_en, Types::String.optional
    attribute :twitter_username, Types::String.optional
    attribute :twitter_hashtag, Types::String.optional
    attribute :syobocal_tid, Types::String.optional
    attribute :mal_anime_id, Types::String.optional
    attribute :is_no_episodes, Types::Bool
    attribute :synopsis_html, Types::String.optional
    attribute :synopsis_en_html, Types::String.optional
    attribute :synopsis_source, Types::String.optional
    attribute :synopsis_source_en, Types::String.optional
    attribute :copyright, Types::String.optional
    attribute :image_url_1x, Types::String.optional
    attribute :image_url_2x, Types::String.optional
    attribute :trailers, Types::Array.of(TrailerEntity)
    attribute :casts, Types::Array.of(CastEntity)
    attribute :staffs, Types::Array.of(StaffEntity)
    attribute :episodes, Types::Array.of(EpisodeEntity)
    attribute :programs, Types::Array.of(ProgramEntity)
    attribute :work_records, Types::Array.of(WorkRecordEntity)
    attribute :series_list, Types::Array.of(SeriesEntity)

    def local_synopsis_html
      return synopsis_en_html if I18n.locale == :en

      synopsis_html
    end

    def local_season_name
      return I18n.t("resources.season.no_season") if season_year.nil? && season_type.nil?

      I18n.t("resources.season.yearly.#{season_type.presence || 'all'}", year: season_year)
    end
  end
end
