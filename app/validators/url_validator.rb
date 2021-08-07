# frozen_string_literal: true

class UrlValidator < ActiveModel::EachValidator
  def validate_each(record, attribute, value)
    return if value.blank?

    unless valid_uri?(value)
      record.errors.add(attribute, :url)
    end
  end

  private

  def valid_uri?(value)
    uri = Addressable::URI.parse(value)
    uri.is_a?(Addressable::URI) && !uri.host.nil?
  rescue Addressable::URI::InvalidURIError
    false
  end
end
