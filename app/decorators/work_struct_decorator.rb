# frozen_string_literal: true

module WorkStructDecorator
  def media_text
    I18n.t("enumerize.work.media.#{media.downcase}")
  end

  def release_season
    return "" if season.blank?
    season.local_name
  end

  def release_season_link
    return release_season if season.blank?
    link_to release_season, season_works_path(season.slug)
  end

  def local_synopsis(raw: false)
    text = case I18n.locale
    when :en then synopsis_en
    else synopsis
    end

    return if text.blank?

    raw ? text : simple_format(text)
  end

  def started_on_label
    case media.downcase
    when "tv" then I18n.t("noun.start_to_broadcast_tv_date")
    when "ova" then I18n.t("noun.start_to_sell_date")
    when "movie" then I18n.t("noun.start_to_broadcast_movie_date")
    else
      I18n.t("noun.start_to_publish_date")
    end
  end

  def twitter_username_link
    return "" if twitter_username.blank?
    link_to "@#{twitter_username}", twitter_username_url, target: "_blank", rel: "noopener"
  end

  def twitter_hashtag_link
    return "" if twitter_hashtag.blank?
    link_to "##{twitter_hashtag}", twitter_hashtag_url, target: "_blank", rel: "noopener"
  end

  def syobocal_link(title = nil)
    return "" if syobocal_tid.blank?
    title = title.presence || syobocal_tid
    link_to title, syobocal_url, target: "_blank", rel: "noopener"
  end

  def mal_anime_link(title = nil)
    return "" if mal_anime_id.blank?
    title = title.presence || mal_anime_id
    link_to title, mal_anime_url, target: "_blank", rel: "noopener"
  end
end
