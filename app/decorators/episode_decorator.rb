# frozen_string_literal: true

class EpisodeDecorator < ApplicationDecorator
  def db_detail_link(options = {})
    name = options.delete(:title).presence || title
    h.link_to name, h.edit_db_episode_path(self), options
  end

  def local_title
    return title if I18n.locale == :ja
    return title_en if title_en.present?
    return title_ro if title_ro.present?
    title
  end

  def local_number
    return number if I18n.locale == :ja
    return "##{raw_number}" if raw_number.present?
    number
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :prev_episode_id
        if send(field).present?
          episode = work.episodes.where(id: send(field)).first
          if episode.present?
            title = episode.decorate.title_with_number
            path = h.work_episode_path(episode.work, episode)
            h.link_to(title, path, target: "_blank")
          else
            ""
          end
        end.to_s
      else
        send(field).to_s
      end

      hash
    end
  end

  def title_with_number
    if local_number.present?
      if local_title.present?
        "#{local_number} #{local_title}"
      else
        local_number
      end
    else
      local_title
    end
  end
end
