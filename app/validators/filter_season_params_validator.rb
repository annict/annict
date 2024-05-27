# typed: false
# frozen_string_literal: true

class FilterSeasonParamsValidator < ActiveModel::EachValidator
  def validate_each(record, _attribute, value)
    return if value.blank?
    regex = /\A[0-9]{4}-(all|spring|summer|autumn|winter)\z/
    record.errors.add(:filter_season, "に不正な値が入っています") unless value&.match?(regex)
  end
end
