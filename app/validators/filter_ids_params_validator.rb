# frozen_string_literal: true

class FilterIdsParamsValidator < ActiveModel::EachValidator
  def validate_each(record, _attribute, value)
    return if value.blank?
    record.errors.add(:filter_ids, "に不正な値が入っています") unless /\A[0-9,]+\z/.match?(value)
  end
end
