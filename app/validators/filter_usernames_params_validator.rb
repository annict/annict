# typed: false
# frozen_string_literal: true

class FilterUsernamesParamsValidator < ActiveModel::EachValidator
  def validate_each(record, _attribute, value)
    return if value.blank?
    return if value.match?(/\A[A-Za-z0-9_,]+\z/)

    message = I18n.t("messages._validators.is_invalid")
    record.errors.add(:filter_usernames, message)
  end
end
