module EpisodeCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(number sort_number title prev_episode_id)
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    validates :number, presence: true
    validates :sort_number, presence: true, numericality: { only_integer: true }

    def to_diffable_hash
      self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = send(field)
        hash
      end
    end
  end
end
