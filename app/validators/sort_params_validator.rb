# frozen_string_literal: true

class SortParamsValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?
    record.errors.add(attribute, "に不正な値が入っています") unless value =~ /\A(asc|desc)\z/
  end
end
