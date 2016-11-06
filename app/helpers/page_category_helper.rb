# frozen_string_literal: true

module PageCategoryHelper
  def page_category
    @page_category.presence || "other"
  end

  # rubocop:disable AccessorMethodName
  def set_page_category(name)
    @page_category = name
    gon.push(basic: { pageCategory: name })
  end
  # rubocop:enable AccessorMethodName
end
