# frozen_string_literal: true

class WorkEntity < ApplicationEntity
  local_attributes :title, :title_alter, :synopsis, :synopsis_source

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

    if database_id = work_node["databaseId"]
      attrs[:database_id] = database_id
    end

    if title = work_node["title"]
      attrs[:title] = title
    end

    if title_en = work_node["titleEn"]
      attrs[:title_en] = title_en
    end

    if title_kana = work_node["titleKana"]
      attrs[:title_kana] = title_kana
    end

    if title_alter = work_node["titleAlter"]
      attrs[:title_alter] = title_alter
    end

    if title_alter_en = work_node["titleAlterEn"]
      attrs[:title_alter_en] = title_alter_en
    end

    if media = work_node["media"]
      attrs[:media] = media.downcase
    end

    if season_year = work_node["seasonYear"]
      attrs[:season_year] = season_year
    end

    if season_type = work_node["seasonType"]
      attrs[:season_type] = season_type.downcase
    end

    if season_slug = work_node["seasonSlug"]
      attrs[:season_slug] = season_slug
    end

    if started_on = work_node["startedOn"]
      attrs[:started_on] = started_on
    end

    if episodes_count = work_node["episodesCount"]
      attrs[:episodes_count] = episodes_count
    end

    if watchers_count = work_node["watchersCount"]
      attrs[:watchers_count] = watchers_count
    end

    if satisfaction_rate = work_node["satisfactionRate"]
      attrs[:satisfaction_rate] = satisfaction_rate
    end

    if ratings_count = work_node["ratingsCount"]
      attrs[:ratings_count] = ratings_count
    end

    if work_records_with_body_count = work_node["workRecordsWithBodyCount"]
      attrs[:work_records_with_body_count] = work_records_with_body_count
    end

    if official_site_url = work_node["officialSiteUrl"]
      attrs[:official_site_url] = official_site_url
    end

    if official_site_url_en = work_node["officialSiteUrlEn"]
      attrs[:official_site_url_en] = official_site_url_en
    end

    if wikipedia_url = work_node["wikipediaUrl"]
      attrs[:wikipedia_url] = wikipedia_url
    end

    if wikipedia_url_en = work_node["wikipediaUrlEn"]
      attrs[:wikipedia_url_en] = wikipedia_url_en
    end

    if twitter_username = work_node["twitterUsername"]
      attrs[:twitter_username] = twitter_username
    end

    if twitter_hashtag = work_node["twitterHashtag"]
      attrs[:twitter_hashtag] = twitter_hashtag
    end

    if syobocal_tid = work_node["syobocalTid"]
      attrs[:syobocal_tid] = syobocal_tid
    end

    if mal_anime_id = work_node["malAnimeId"]
      attrs[:mal_anime_id] = mal_anime_id
    end

    if is_no_episodes = work_node["isNoEpisodes"]
      attrs[:is_no_episodes] = is_no_episodes
    end

    if synopsis = work_node["synopsis"]
      attrs[:synopsis] = synopsis
    end

    if synopsis_en = work_node["synopsisEn"]
      attrs[:synopsis_en] = synopsis_en
    end

    if synopsis_source = work_node["synopsisSource"]
      attrs[:synopsis_source] = synopsis_source
    end

    if synopsis_source_en = work_node["synopsisSourceEn"]
      attrs[:synopsis_source_en] = synopsis_source_en
    end

    if copyright = work_node["copyright"]
      attrs[:copyright] = copyright
    end

    if image_url_1x = work_node.dig("image", "internalUrl1x")
      attrs[:image_url_1x] = image_url_1x
    end

    if image_url_2x = work_node.dig("image", "internalUrl2x")
      attrs[:image_url_2x] = image_url_2x
    end

    trailer_nodes = work_node.dig("trailers", "nodes")
    if trailer_nodes.present?
      attrs[:trailers] = TrailerEntity.from_nodes(trailer_nodes)
    end

    cast_nodes = work_node.dig("casts", "nodes")
    if cast_nodes.present?
      attrs[:casts] = CastEntity.from_nodes(cast_nodes)
    end

    staff_nodes = work_node.dig("staffs", "nodes")
    if staff_nodes.present?
      attrs[:staffs] = StaffEntity.from_nodes(staff_nodes)
    end

    episode_nodes = work_node.dig("episodes", "nodes")
    if episode_nodes.present?
      attrs[:episodes] = EpisodeEntity.from_nodes(episode_nodes)
    end

    program_nodes = work_node.dig("programs", "nodes")
    if program_nodes.present?
      attrs[:programs] = ProgramEntity.from_nodes(program_nodes)
    end

    work_record_nodes = work_node.dig("workRecords", "nodes")
    if work_record_nodes.present?
      attrs[:work_records] = WorkRecordEntity.from_nodes(work_record_nodes)
    end

    series_nodes = work_node.dig("seriesList", "nodes")
    if series_nodes.present?
      attrs[:series_list] = SeriesEntity.from_nodes(series_nodes)
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
