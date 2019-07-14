# frozen_string_literal: true

class ApplicationStruct < Dry::Struct
  transform_types do |type|
    type.required(false)
  end

  def decorate
    ActiveDecorator::Decorator.instance.decorate(self)
  end

  def method_missing(method_name, *arguments, &block)
    return super if method_name.blank?
    return super unless method_name.to_s.start_with?("local_")
    local_attribute(method_name.to_s.sub("local_", ""), *arguments)
  end

  def respond_to_missing?(method_name, include_private = false)
    method_name.to_s.start_with?("local_") || super
  end

  private

  def local_attribute(attribute_name, fallback: true)
    property_ja = send(attribute_name.to_sym)
    property_en = send("#{attribute_name}_en".to_sym)

    return property_ja if I18n.locale == :ja
    return property_en if property_en.present?

    property_ja if fallback
  end
end
