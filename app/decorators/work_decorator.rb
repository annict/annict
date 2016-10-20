# frozen_string_literal: true

class WorkDecorator < ApplicationDecorator
  include RootResourceDecoratorCommon

  def twitter_username_link
    url = "https://twitter.com/#{twitter_username}"
    h.link_to "@#{twitter_username}", url, target: "_blank"
  end

  def twitter_hashtag_link
    url = URI.encode("https://twitter.com/search?q=##{twitter_hashtag}&src=hash")
    h.link_to "##{twitter_hashtag}", url, target: "_blank"
  end

  def syobocal_link(title = nil)
    title = title.presence || sc_tid
    h.link_to title, syobocal_url, target: "_blank"
  end

  def db_detail_link(options = {})
    name = options.delete(:name).presence || title
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_path(self)
    else
      h.new_db_draft_work_path(work_id: id)
    end

    h.link_to name, path, options
  end

  def release_season
    return "" if season.blank?
    season.decorate.humanize_name
  end

  def release_season_link
    return release_season if season.blank?
    h.link_to release_season, h.season_works_path(season.slug)
  end

  def db_header_title
    local_title
  end

  def local_title
    return title if I18n.locale == :ja
    return title_en if title_en.present?
    return title_ro if title_ro.present?
    title
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :season_id
        if send(:season_id).present?
          Season.find(send(:season_id)).decorate.yearly_season_ja
        end
      when :sc_tid
        sc_tid = send(:sc_tid)
        if sc_tid.present?
          url = "http://cal.syoboi.jp/tid/#{sc_tid}"
          h.link_to(sc_tid, url, target: "_blank")
        end
      when :media
        Work.media.find_value(send(:media)).text
      when :official_site_url, :wikipedia_url
        url = send(field)
        if url.present?
          begin
            h.link_to(URI.decode(url), url, target: "_blank")
          rescue
            url
          end
        end
      when :twitter_username
        username = send(:twitter_username)
        if username.present?
          url = "https://twitter.com/#{username}"
          h.link_to("@#{username}", url, target: "_blank")
        end
      when :twitter_hashtag
        hashtag = send(:twitter_hashtag)
        if hashtag.present?
          url = "https://twitter.com/search?q=%23#{hashtag}"
          h.link_to("##{hashtag}", url, target: "_blank")
        end
      when :number_format_id
        send(:number_format).name if send(:number_format_id).present?
      else
        send(field)
      end

      hash
    end
  end
end
