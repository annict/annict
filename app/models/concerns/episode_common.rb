module EpisodeCommon
  extend ActiveSupport::Concern

  included do
    DIFF_FIELDS = %i(number sort_number title next_episode_id)

    validates :number, presence: true
    validates :sort_number, presence: true, numericality: { only_integer: true }
    validates :title, presence: true

    def to_diffable_hash
      self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = send(field)
        hash
      end
    end
  end
end
