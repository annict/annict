module Diffable
  def get_diffable_work(column_name, value)
    case column_name.to_s
    when "season_id"
      season = Season.find(value)
      { data: value, value: season.name }
    when "sc_tid"
      path = "http://cal.syoboi.jp/tid/#{value}"
      { data: value, value: h.link_to(value, path, target: "_blank") }
    when "media"
      { data: value.to_s, value: Work.media.find_value(value).text }
    when "official_site_url", "wikipedia_url"
      { data: value, value: h.link_to(value, value, target: "_blank") }
    when "twitter_username"
      path = "https://twitter.com/#{value}"
      { data: value, value: h.link_to(value, path, target: "_blank") }
    when "twitter_hashtag"
      path = "https://twitter.com/search?q=%23#{value}"
      { data: "#{value}", value: h.link_to("#{value}", path, target: "_blank") }
    when "released_at"
      if value.is_a?(String)
        released_at = Date.parse(value).strftime("%Y-%m-%d")
        { data: released_at, value: released_at }
      elsif value.is_a?(Date)
        released_at = value.strftime("%Y-%m-%d")
        { data: released_at, value: released_at }
      end
    else
      { data: value, value: value }
    end
  end

  def get_diffable_episode(column_name, value)
    case column_name.to_s
    when "next_episode_id"
      episode = Episode.find(value)
      title = episode.decorate.title_with_number
      path = h.work_episode_path(episode.work, episode)

      { data: value, value: h.link_to(title, path, target: "_blank") }
    else
      { data: value, value: value }
    end
  end
end
