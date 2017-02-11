# frozen_string_literal: true

class CharacterDecorator < ApplicationDecorator
  include RootResourceDecoratorCommon

  def name_link
    h.link_to local_name, h.character_path(self)
  end

  def db_header_title
    local_name
  end

  def name_with_kind
    return "#{local_name} (#{local_kind})" if local_kind.present?
    local_name
  end

  def grid_description(cast)
    "CV: #{cast.person.decorate.name_link}"
  end

  def to_values
    model.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
      hash[field] = send(field)
      hash
    end
  end

  def db_detail_link(options = {})
    name = options.delete(:name).presence || self.name
    path = h.edit_db_character_path(self)
    h.link_to name, path, options
  end

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
