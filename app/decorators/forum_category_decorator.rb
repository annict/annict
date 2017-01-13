# frozen_string_literal: true

class ForumCategoryDecorator < ApplicationDecorator
  def local_name
    return name if I18n.locale == :ja
    name_en
  end
end
