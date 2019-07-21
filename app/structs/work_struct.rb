# frozen_string_literal: true

class WorkStruct < ApplicationStruct
  attribute :id, StructTypes::Strict::String
  attribute :annict_id, StructTypes::Strict::Integer
  attribute :title, StructTypes::Strict::String
  attribute :title_kana, StructTypes::Strict::String
  attribute :title_en, StructTypes::Strict::String
  attribute :media, StructTypes::Strict::String
  attribute :season_year, StructTypes::Strict::Integer
  attribute :season_name, StructTypes::Strict::String
  attribute :started_on, StructTypes::Strict::String
  attribute :watchers_count, StructTypes::Strict::Integer
  attribute :copyright, StructTypes::Strict::String
  attribute :satisfaction_rate, StructTypes::Strict::Float
  attribute :ratings_count, StructTypes::Strict::Integer
  attribute :official_site_url, StructTypes::Strict::String
  attribute :official_site_url_en, StructTypes::Strict::String
  attribute :wikipedia_url, StructTypes::Strict::String
  attribute :wikipedia_url_en, StructTypes::Strict::String
  attribute :twitter_username, StructTypes::Strict::String
  attribute :twitter_hashtag, StructTypes::Strict::String
  attribute :syobocal_tid, StructTypes::Strict::Integer
  attribute :mal_anime_id, StructTypes::Strict::String
  attribute :is_no_episodes, StructTypes::Strict::Bool
  attribute :synopsis, StructTypes::Strict::String
  attribute :synopsis_en, StructTypes::Strict::String
  attribute :synopsis_source, StructTypes::Strict::String
  attribute :synopsis_source_en, StructTypes::Strict::String
  attribute :viewer_status_state, StructTypes::Strict::String

  attribute :image, WorkImageStruct
  attribute :trailers, StructTypes::Strict::Array.of(TrailerStruct)
  attribute :casts, StructTypes::Strict::Array.of(CastStruct)
  attribute :staffs, StructTypes::Strict::Array.of(StaffStruct)
  attribute :episodes, StructTypes::Strict::Array.of(EpisodeStruct)
  attribute :programs, StructTypes::Strict::Array.of(ProgramStruct)

  def season
    return if season_year.blank?
    @season ||= Season.new(season_year, season_name&.downcase.presence || "all")
  end

  def casted_started_on
    Date.parse(started_on)
  end

  def twitter_username_url
    return "" if twitter_username.blank?
    url = "https://twitter.com/#{twitter_username}"
    Addressable::URI.parse(url).normalize.to_s
  end

  def twitter_hashtag_url
    return "" if twitter_hashtag.blank?
    url = "https://twitter.com/search?q=##{twitter_hashtag}&src=hash"
    Addressable::URI.parse(url).normalize.to_s
  end

  def syobocal_url
    return "" if syobocal_tid.blank?
    "http://cal.syoboi.jp/tid/#{syobocal_tid}"
  end

  def mal_anime_url
    return "" if mal_anime_id.blank?
    "https://myanimelist.net/anime/#{mal_anime_id}"
  end
end
