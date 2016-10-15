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
end
