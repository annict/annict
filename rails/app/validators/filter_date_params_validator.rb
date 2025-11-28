# typed: false
# frozen_string_literal: true

class FilterDateParamsValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?

    begin
      DateTime.parse(value)
    rescue
      record.errors.add(attribute, "の値が不正です")
    end
  end
end
