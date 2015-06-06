module EpisodeDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :next_episode_id
          episode = work.episodes.find(send(field))
          title = episode.decorate.title_with_number
          path = h.work_episode_path(episode.work, episode)
          h.link_to(title, path, target: "_blank")
        else
          send(field)
        end

        hash
      end
    end

    def title_with_number
      if number.present?
        if title.present?
          "#{number}「#{title}」"
        else
          number
        end
      else
        title
      end
    end
  end
end
