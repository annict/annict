# frozen_string_literal: true

module SlotDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    link_to(name, db_edit_slot_path(self), options)
  end

  def state_text
    I18n.t("noun.#{state}")
  end

  def to_values
    self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :channel_id
        Channel.find(send(field)).name
      when :episode_id
        next unless send(field)

        episode = anime.episodes.find(send(field))
        title = episode.decorate.title_with_number
        path = episode_path(anime_id: episode.work_id, episode_id: episode.id)
        link_to(title, path, target: "_blank")
      when :work_id
        path = anime_path(anime_id: anime.id)
        link_to(anime.title, path, target: "_blank")
      when :started_at
        send(field).in_time_zone("Asia/Tokyo").strftime("%Y/%m/%d %H:%M")
      when :rebroadcast
        send(field) ? icon("check") : "-"
      else
        send(field)
      end

      hash
    end
  end
end
