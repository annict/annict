# typed: false
# frozen_string_literal: true

module SeriesDecorator
  include RootResourceDecoratorCommon

  def local_name
    return name if I18n.locale == :ja
    return name_ro if name_ro.present?
    return name_en if name_en.present?
    name
  end
end
