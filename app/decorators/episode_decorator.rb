# frozen_string_literal: true

module EpisodeDecorator
  def number_link
    link_to local_number, work_episode_path(work, self)
  end

  def db_detail_link(options = {})
    name = options.delete(:title).presence || title
    link_to name, edit_db_episode_path(self), options
  end

  def local_title(fallback: true)
    return title if I18n.locale == :ja
    return title_en if title_en.present?
    title if fallback
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :prev_episode_id
        if send(field).present?
          episode = work.episodes.where(id: send(field)).first
          if episode.present?
            title = episode.title_with_number
            path = work_episode_path(episode.work, episode)
            link_to(title, path, target: "_blank")
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

  def title_with_number(fallback: true)
    l_number = local_number
    l_title = local_title(fallback: fallback)

    return l_title if l_number.blank?
    return "#{l_number} #{l_title}" if l_title.present?

    l_number
  end

  def number_with_work_title
    work_title = work.local_title

    return work_title if work.single?
    return "#{work_title} #{local_title}" if local_number.blank?

    "#{work_title} #{local_number}"
  end

  def delete_item_path(item)
    episode_item_path(self, item)
  end
end
