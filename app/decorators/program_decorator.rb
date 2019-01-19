# frozen_string_literal: true

module ProgramDecorator
  def db_detail_link(options = {})
    name = options.delete(:name).presence || id
    h.link_to(name, h.edit_db_program_path(self), options)
  end

  def state_text
    I18n.t("noun.#{state}")
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = case field
      when :channel_id
        Channel.find(send(field)).name
      when :episode_id
        episode = work.episodes.find(send(field))
        title = episode.decorate.title_with_number
        path = h.work_episode_path(episode.work, episode)
        h.link_to(title, path, target: "_blank")
      when :work_id
        path = h.work_path(work)
        h.link_to(work.title, path, target: "_blank")
      when :started_at
        send(field).in_time_zone("Asia/Tokyo").strftime("%Y/%m/%d %H:%M")
      when :rebroadcast
        send(field) ? h.icon("check") : "-"
      else
        send(field)
      end

      hash
    end
  end
end
