# frozen_string_literal: true

class LocaleValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.match?(/\A(ja|en)\z/)
    record.errors.add(attribute, "に不正な値が入っています")
  end
end
