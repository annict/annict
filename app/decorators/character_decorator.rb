# frozen_string_literal: true

class CharacterDecorator < ApplicationDecorator
  include RootResourceDecoratorCommon

  def db_header_title
    local_name
  end

  def local_name
    return name if I18n.locale == :ja
    return name_en if name_en.present?
    name
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
end
