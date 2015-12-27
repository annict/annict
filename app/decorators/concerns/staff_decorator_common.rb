module StaffDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :work_id
          work_id = send(:work_id)
          Work.find(work_id).title if work_id.present?
        when :role
          send(:role_text)
        else
          send(field)
        end

        hash
      end
    end

    def role_name
      return role_other if role_value == "other"
      role_text
    end
  end
end
