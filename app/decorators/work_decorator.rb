# frozen_string_literal: true

module WorkDecorator
  include RootResourceDecoratorCommon

  def title_link
    link_to local_title, work_path(self)
  end

  def twitter_username_link
    link_to "@#{twitter_username}", twitter_username_url, target: "_blank"
  end

  def twitter_hashtag_link
    link_to "##{twitter_hashtag}", twitter_hashtag_url, target: "_blank"
  end

  def syobocal_link(title = nil)
    title = title.presence || sc_tid
    link_to title, syobocal_url, target: "_blank"
  end

  def mal_anime_link(title = nil)
    title = title.presence || mal_anime_id
    link_to title, mal_anime_url, target: "_blank"
  end

  def db_detail_link(options = {})
    name = options.delete(:name).presence || title
    link_to(name, edit_db_work_path(self), options)
  end

  def release_season
    return "" if season.blank?
    season.local_name
  end

  def release_season_link
    return release_season if season.blank?
    link_to release_season, season_works_path(season.slug)
  end

  def db_header_title
    local_title
  end

  def local_synopsis(raw: false)
    text = case I18n.locale
    when :ja then synopsis
    when :en then synopsis_en
    end

    return "" if text.blank?

    raw ? text : simple_format(text)
  end

  def media_label
    content_tag :span, class: "badge u-badge-works" do
      media.text
    end
  end

  def started_on_label
    if media.tv?
      I18n.t("noun.start_to_broadcast_tv_date")
    elsif media.ova?
      I18n.t("noun.start_to_sell_date")
    elsif media.movie?
      I18n.t("noun.start_to_broadcast_movie_date")
    else
      I18n.t("noun.start_to_publish_date")
    end
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :sc_tid
        sc_tid = send(:sc_tid)
        if sc_tid.present?
          url = "http://cal.syoboi.jp/tid/#{sc_tid}"
          link_to(sc_tid, url, target: "_blank")
        end
      when :media
        Work.media.find_value(send(:media)).text
      when :official_site_url, :wikipedia_url
        url = send(field)
        if url.present?
          begin
            link_to(URI.decode(url), url, target: "_blank")
          rescue
            url
          end
        end
      when :twitter_username
        username = send(:twitter_username)
        if username.present?
          url = "https://twitter.com/#{username}"
          link_to("@#{username}", url, target: "_blank")
        end
      when :twitter_hashtag
        hashtag = send(:twitter_hashtag)
        if hashtag.present?
          url = "https://twitter.com/search?q=%23#{hashtag}"
          link_to("##{hashtag}", url, target: "_blank")
        end
      when :number_format_id
        send(:number_format).name if send(:number_format_id).present?
      when :season_year
        send(:season_year).to_s
      when :season_name
        send(:season_name)&.text
      when :started_on
        send(:started_on).to_s
      when :ended_on
        send(:ended_on).to_s
      else
        send(field)
      end

      hash
    end
  end
end
