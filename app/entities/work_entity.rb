# frozen_string_literal: true

class WorkEntity < ApplicationEntity
  local_attributes :title, :title_alter, :synopsis, :synopsis_source

  attribute? :id, Types::Integer
  attribute? :database_id, Types::Integer
  attribute? :title, Types::String
  attribute? :title_en, Types::String.optional
  attribute? :title_kana, Types::String.optional
  attribute? :title_alter, Types::String.optional
  attribute? :title_alter_en, Types::String.optional
  attribute? :media, Types::WorkMediaKinds
  attribute? :season_year, Types::Integer.optional
  attribute? :season_type, Types::SeasonKinds.optional
  attribute? :season_slug, Types::String.optional
  attribute? :started_on, Types::Params::Date.optional
  attribute? :episodes_count, Types::Integer
  attribute? :watchers_count, Types::Integer
  attribute? :satisfaction_rate, Types::Float.optional
  attribute? :ratings_count, Types::Integer
  attribute? :work_records_with_body_count, Types::Integer
  attribute? :official_site_url, Types::String.optional
  attribute? :official_site_url_en, Types::String.optional
  attribute? :wikipedia_url, Types::String.optional
  attribute? :wikipedia_url_en, Types::String.optional
  attribute? :twitter_username, Types::String.optional
  attribute? :twitter_hashtag, Types::String.optional
  attribute? :syobocal_tid, Types::String.optional
  attribute? :mal_anime_id, Types::String.optional
  attribute? :is_no_episodes, Types::Bool
  attribute? :synopsis, Types::String.optional
  attribute? :synopsis_en, Types::String.optional
  attribute? :synopsis_html, Types::String.optional
  attribute? :synopsis_en_html, Types::String.optional
  attribute? :synopsis_source, Types::String.optional
  attribute? :synopsis_source_en, Types::String.optional
  attribute? :copyright, Types::String.optional
  attribute? :image_url_1x, Types::String.optional
  attribute? :image_url_2x, Types::String.optional
  attribute? :trailers, Types::Array.of(TrailerEntity)
  attribute? :casts, Types::Array.of(CastEntity)
  attribute? :staffs, Types::Array.of(StaffEntity)
  attribute? :episodes, Types::Array.of(EpisodeEntity)
  attribute? :programs, Types::Array.of(ProgramEntity)
  attribute? :work_records, Types::Array.of(WorkRecordEntity)
  attribute? :series_list, Types::Array.of(SeriesEntity)

  def self.from_model(work)
    new(
      media: work.media.to_s,
      syobocal_tid: work.sc_tid&.to_s,
      mal_anime_id: work.mal_anime_id&.to_s
    )
  end

  def self.from_node(work_node)
    attrs = {}

    if database_id = work_node["annictId"]
      attrs[:database_id] = database_id
    end

    if title = work_node["title"]
      attrs[:title] = title
    end

    if title_en = work_node["titleEn"]
      attrs[:title_en] = title_en
    end

    if image_url_1x = work_node.dig("image", "internalUrl1x")
      attrs[:image_url_1x] = image_url_1x
    end

    if image_url_2x = work_node.dig("image", "internalUrl2x")
      attrs[:image_url_2x] = image_url_2x
    end

    new attrs
  end

  def local_season_name
    return I18n.t("resources.season.no_season") if season_year.nil? && season_type.nil?

    I18n.t("resources.season.yearly.#{season_type.presence || 'all'}", year: season_year)
  end

  def media_text
    I18n.t("enumerize.work.media.#{media}")
  end

  def twitter_profile_url
    url = "https://twitter.com/#{twitter_username}"
    Addressable::URI.parse(url).normalize.to_s
  end

  def twitter_hashtag_url
    url = "https://twitter.com/search?q=%23#{twitter_hashtag}"
    Addressable::URI.parse(url).normalize.to_s
  end

  def syobocal_url
    "http://cal.syoboi.jp/tid/#{syobocal_tid}"
  end

  def mal_anime_url
    "https://myanimelist.net/anime/#{mal_anime_id}"
  end
end
