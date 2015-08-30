module EpisodeCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(number sort_number title prev_episode_id)
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    validates :sort_number, presence: true, numericality: { only_integer: true }

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = send(field) if send(field).present?
        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
