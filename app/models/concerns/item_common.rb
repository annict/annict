module ItemCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(name url tombo_image main)
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    has_attached_file :tombo_image

    validates :name, presence: true
    validates :url, presence: true, url: true, amazon: true
    validates :tombo_image, attachment_presence: true,
                            attachment_content_type: { content_type: /\Aimage/ }

    after_save :switch_main_flag

    def to_diffable_hash
      self.class::DIFF_FIELDS.inject({}) do |hash, field|
        hash[field] = case field
        when :tombo_image
          send(field).size
        else
          send(field)
        end

        hash
      end.delete_if { |_, v| v.blank? }
    end

    private

    def switch_main_flag
      if main?
        work.items.where.not(id: id).update_all(main: false)
      end
    end
  end
end
