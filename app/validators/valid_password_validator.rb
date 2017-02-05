# frozen_string_literal: true

class ValidPasswordValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?
    return if record.valid_password?(value)
    record.errors.add(attribute, I18n.t("resources.user.is_wrong"))
  end
end
