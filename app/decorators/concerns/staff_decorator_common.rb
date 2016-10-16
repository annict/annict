# frozen_string_literal: true

module StaffDecoratorCommon
  extend ActiveSupport::Concern

  included do
    def to_values
      model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :resource_id
          resource&.name
        when :role
          send(:role_text)
        when :sort_number
          send(:sort_number).to_s
        else
          send(field)
        end

        hash
      end
    end

    def role_name
      return i18n_role_other if role_value == "other"
      role_text
    end

    def i18n_role_other
      return role_other_en if I18n.locale != :ja && role_other_en.present?
      role_other
    end
  end
end
