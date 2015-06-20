module ProgramCommon
  extend ActiveSupport::Concern

  included do
    DIFF_FIELDS = %i(channel_id episode_id started_at)
    PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

    belongs_to :channel
    belongs_to :episode
    belongs_to :work

    validates :channel_id, presence: true
    validates :episode_id, presence: true
    validates :started_at, presence: true

    def to_diffable_hash
      self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = send(field)
        hash
      end
    end
  end
end
