# frozen_string_literal: true

module UserCheckable
  extend ActiveSupport::Concern

  included do
    def tracked?(episode)
      records.exists?(episode_id: episode.id)
    end

    def records_count_in(episode)
      records.where(episode: episode).count
    end
  end
end
