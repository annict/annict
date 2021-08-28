# frozen_string_literal: true

module Forms
  class EpisodeRecordForm < Forms::ApplicationForm
    include Forms::Recordable

    attr_accessor :advanced_rating, :episode

    validates :advanced_rating, allow_nil: true, numericality: {greater_than_or_equal_to: 1, less_than_or_equal_to: 5}
    validates :episode, presence: true

    def episode
      @episode ||= Episode.only_kept.find_by(id: episode_id)
    end
  end
end
