# frozen_string_literal: true

module V6::PageCategorizable
  private

  def page_category
    @page_category ||= "other"
  end

  def set_page_category(name)
    @page_category = name
  end
end
