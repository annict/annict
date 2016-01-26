module ProgramCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(channel_id episode_id started_at rebroadcast)
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    belongs_to :channel
    belongs_to :episode
    belongs_to :work, touch: true

    validates :channel_id, presence: true
    validates :episode_id, presence: true
    validates :started_at, presence: true

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = send(field)
        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
