# frozen_string_literal: true

class FieldsParamsValidator < ActiveModel::EachValidator
  def validate_each(record, _attribute, value)
    return if value.blank?
    record.errors.add(:fields, "に不正な値が入っています") unless value =~ /\A[a-z0-9_.,]+\z/
  end
end
