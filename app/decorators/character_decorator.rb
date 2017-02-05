# frozen_string_literal: true

class CharacterDecorator < ApplicationDecorator
  include RootResourceDecoratorCommon

  def name_link
    h.link_to local_name, h.character_path(self)
  end

  def db_header_title
    local_name
  end

  def local_name
    return name if I18n.locale == :ja
    return name_en if name_en.present?
    name
  end

  def local_kind
    return kind if I18n.locale == :ja
    return kind_en if kind_en.present?
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
end
