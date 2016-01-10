module StaffDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :person_id
          person_id = send(:person_id)
          Person.find(person_id).name if person_id.present?
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
