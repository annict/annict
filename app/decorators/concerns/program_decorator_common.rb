module ProgramDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.inject({}) do |hash, field|
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
          send(field).to_time.strftime("%Y/%m/%d %H:%M")
        when :rebroadcast
          send(field) ? h.icon("check") : "-"
        else
          send(field)
        end

        hash
      end
    end
  end
end
