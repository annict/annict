# frozen_string_literal: true

class AsinValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?
    return if value =~ /\A[0-9A-Z]{10}\z/
    record.errors.add(attribute, I18n.t("resources.work_image.invalid_asin"))
  end
end
