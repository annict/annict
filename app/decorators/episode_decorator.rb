# frozen_string_literal: true

class EpisodeDecorator < ApplicationDecorator
  include EpisodeDecoratorCommon

  def db_detail_link(options = {})
    name = options.delete(:name).presence || name
    path = if h.user_signed_in? && h.current_user.committer?
      h.edit_db_work_episode_path(work, self)
    else
      h.new_db_work_draft_episode_path(work, id: id)
    end

    h.link_to name, path, options
  end

  def meta_number(prefix: false)
    result = raw_number.present? ? "第#{raw_number}話" : (number.presence || "")
    return "- #{result}" if result.present? && prefix
    result
  end

  def meta_title
    return "" if title.blank? || title == work.title
    "「#{title}」"
  end

  def formatted_number
    return number if I18n.locale == :ja
    return "##{raw_number}" if raw_number.present?
    number
  end
end
