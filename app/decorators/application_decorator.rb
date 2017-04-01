# frozen_string_literal: true

class ApplicationDecorator < Draper::Decorator
  delegate_all

  def method_missing(method_name, *arguments, &block)
    return super if method_name.blank?
    return super unless method_name.to_s.start_with?("local_")
    _local_property(method_name.to_s.sub("local_", ""))
  end

  def respond_to_missing?(method_name, include_private = false)
    method_name.to_s.start_with?("local_") || super
  end

  private

  def _local_property(property_name)
    property_ja = send(property_name.to_sym)
    property_en = send("#{property_name}_en".to_sym)

    return property_ja if I18n.locale == :ja
    return property_en if property_en.present?

    property_ja
  end
end
