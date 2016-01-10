class WorkDecorator < ApplicationDecorator
  include WorkDecoratorCommon

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
    return season.decorate.yearly_season_ja if season.present?
    released_at_about.presence || ""
  end
end
