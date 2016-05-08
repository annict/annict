# frozen_string_literal: true

class FilterBooleanParamsValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?
    record.errors.add(attribute, "の値が不正です") unless value =~ /\A(true|false)\z/
  end
end
