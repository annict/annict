# typed: false
# frozen_string_literal: true

module UserCheckable
  extend ActiveSupport::Concern

  included do
    def tracked?(episode)
      episode_records.exists?(episode_id: episode.id)
    end

    def episode_records_count_in(episode)
      episode_records.where(episode: episode).count
    end
  end
end
