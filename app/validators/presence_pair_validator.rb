# typed: false
# frozen_string_literal: true

class PresencePairValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    pair_column_name = options[:with]
    pair_value = record.send(pair_column_name)

    return if value.blank? && pair_value.blank?
    return if value.present? && pair_value.present?

    i18n_message_key = "messages._validators.and_other_are_required"
    i18n_column_name =
      I18n.t("activerecord.attributes.#{record.class.name.downcase}.#{pair_column_name}")
    record.errors.add(attribute, I18n.t(i18n_message_key, column_name: i18n_column_name))
  end
end
