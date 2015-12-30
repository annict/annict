module WorkOrganizationCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(organization_id role role_other)
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    extend Enumerize

    enumerize :role, in: %w(
      producer
      other
    )

    belongs_to :organization
    belongs_to :work

    validates :organization_id, presence: true
    validates :work_id, presence: true
    validates :role, presence: true

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :role
          send(field).to_s if send(field).present?
        else
          send(field)
        end

        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
