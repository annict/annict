# typed: false
# frozen_string_literal: true

class SortParamsValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?
    record.errors.add(attribute, "に不正な値が入っています") unless /\A(asc|desc)\z/.match?(value)
  end
end
