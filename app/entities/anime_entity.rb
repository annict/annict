# frozen_string_literal: true

class AnimeEntity < ApplicationEntity
  local_attributes :title, :title_alter, :synopsis, :synopsis_source

  attribute? :id, Types::String
  attribute? :database_id, Types::Integer
  attribute? :title, Types::String
  attribute? :title_en, Types::String.optional
  attribute? :title_kana, Types::String.optional
  attribute? :title_alter, Types::String.optional
  attribute? :title_alter_en, Types::String.optional
  attribute? :media, Types::AnimeMediaKinds
  attribute? :season_year, Types::Integer.optional
  attribute? :season_type, Types::SeasonKinds.optional
  attribute? :season_slug, Types::String.optional
  attribute? :started_on, Types::Params::Date.optional
  attribute? :episodes_count, Types::Integer
  attribute? :final_episodes_count, Types::Integer
  attribute? :watchers_count, Types::Integer
  attribute? :satisfaction_rate, Types::Float.optional
  attribute? :ratings_count, Types::Integer
  attribute? :anime_records_with_body_count, Types::Integer
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
  attribute? :records, Types::Array.of(RecordEntity)
  attribute? :series_list, Types::Array.of(SeriesEntity)

  def self.from_model(work)
    new(
      media: work.media.to_s,
      syobocal_tid: work.sc_tid&.to_s,
      mal_anime_id: work.mal_anime_id&.to_s
    )
  end

  def self.from_node(node)
    attrs = {}

    if id = node["id"]
      attrs[:id] = id
    end

    if database_id = node["databaseId"]
      attrs[:database_id] = database_id
    end

    if title = node["title"]
      attrs[:title] = title
    end

    if title_en = node["titleEn"]
      attrs[:title_en] = title_en
    end

    if title_kana = node["titleKana"]
      attrs[:title_kana] = title_kana
    end

    if title_alter = node["titleAlter"]
      attrs[:title_alter] = title_alter
    end

    if title_alter_en = node["titleAlterEn"]
      attrs[:title_alter_en] = title_alter_en
    end

    if media = node["media"]
      attrs[:media] = media.downcase
    end

    if season_year = node["seasonYear"]
      attrs[:season_year] = season_year
    end

    if season_type = node["seasonType"]
      attrs[:season_type] = season_type.downcase
    end

    if season_slug = node["seasonSlug"]
      attrs[:season_slug] = season_slug
    end

    if started_on = node["startedOn"]
      attrs[:started_on] = started_on
    end

    if episodes_count = node["episodesCount"]
      attrs[:episodes_count] = episodes_count
    end

    if final_episodes_count = node["finalEpisodesCount"]
      attrs[:final_episodes_count] = final_episodes_count
    end

    if watchers_count = node["watchersCount"]
      attrs[:watchers_count] = watchers_count
    end

    if satisfaction_rate = node["satisfactionRate"]
      attrs[:satisfaction_rate] = satisfaction_rate
    end

    if ratings_count = node["ratingsCount"]
      attrs[:ratings_count] = ratings_count
    end

    if anime_records_with_body_count = node["animeRecordsWithBodyCount"]
      attrs[:anime_records_with_body_count] = anime_records_with_body_count
    end

    if official_site_url = node["officialSiteUrl"]
      attrs[:official_site_url] = official_site_url
    end

    if official_site_url_en = node["officialSiteUrlEn"]
      attrs[:official_site_url_en] = official_site_url_en
    end

    if wikipedia_url = node["wikipediaUrl"]
      attrs[:wikipedia_url] = wikipedia_url
    end

    if wikipedia_url_en = node["wikipediaUrlEn"]
      attrs[:wikipedia_url_en] = wikipedia_url_en
    end

    if twitter_username = node["twitterUsername"]
      attrs[:twitter_username] = twitter_username
    end

    if twitter_hashtag = node["twitterHashtag"]
      attrs[:twitter_hashtag] = twitter_hashtag
    end

    if syobocal_tid = node["syobocalTid"]
      attrs[:syobocal_tid] = syobocal_tid
    end

    if mal_anime_id = node["malAnimeId"]
      attrs[:mal_anime_id] = mal_anime_id
    end

    if is_no_episodes = node["isNoEpisodes"]
      attrs[:is_no_episodes] = is_no_episodes
    end

    if synopsis = node["synopsis"]
      attrs[:synopsis] = synopsis
    end

    if synopsis_en = node["synopsisEn"]
      attrs[:synopsis_en] = synopsis_en
    end

    if synopsis_source = node["synopsisSource"]
      attrs[:synopsis_source] = synopsis_source
    end

    if synopsis_source_en = node["synopsisSourceEn"]
      attrs[:synopsis_source_en] = synopsis_source_en
    end

    if copyright = node["copyright"]
      attrs[:copyright] = copyright
    end

    if image_url_1x = node.dig("image", "internalUrl1x")
      attrs[:image_url_1x] = image_url_1x
    end

    if image_url_2x = node.dig("image", "internalUrl2x")
      attrs[:image_url_2x] = image_url_2x
    end

    trailer_nodes = node.dig("trailers", "nodes")
    attrs[:trailers] = TrailerEntity.from_nodes(trailer_nodes || [])

    cast_nodes = node.dig("casts", "nodes")
    attrs[:casts] = CastEntity.from_nodes(cast_nodes || [])

    staff_nodes = node.dig("staffs", "nodes")
    attrs[:staffs] = StaffEntity.from_nodes(staff_nodes || [])

    episode_nodes = node.dig("episodes", "nodes")
    attrs[:episodes] = EpisodeEntity.from_nodes(episode_nodes || [])

    program_nodes = node.dig("programs", "nodes")
    attrs[:programs] = ProgramEntity.from_nodes(program_nodes || [])

    record_nodes = node.dig("records", "nodes")
    attrs[:records] = RecordEntity.from_nodes(record_nodes || [])

    series_nodes = node.dig("seriesList", "nodes")
    attrs[:series_list] = SeriesEntity.from_nodes(series_nodes || [])

    new attrs
  end

  def local_season_name
    return I18n.t("resources.season.no_season") if season_year.nil? && season_type.nil?

    I18n.t("resources.season.yearly.#{season_type.presence || "all"}", year: season_year)
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
