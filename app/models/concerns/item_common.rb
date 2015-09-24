module ItemCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(name url tombo_image)
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    has_attached_file :tombo_image, preserve_files: true

    validates :name, presence: true
    validates :url, presence: true, url: true, amazon: true, length: { maximum: 500 }
    validates :tombo_image, attachment_presence: true,
                            attachment_content_type: { content_type: /\Aimage/ }

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :tombo_image
          send(field).size
        else
          send(field)
        end

        hash
      end

      data.delete_if { |_, v| v.blank? }
    end
  end
end
