# typed: false
# frozen_string_literal: true

module ForumCategoryDecorator
  def local_name
    return name if I18n.locale == :ja
    name_en
  end
end
