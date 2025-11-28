# typed: false
# frozen_string_literal: true

module PageCategorizable
  extend ActiveSupport::Concern

  included do
    helper_method :page_category
  end

  private

  def page_category
    @page_category ||= "other"
  end

  def set_page_category(name)
    @page_category = name
  end
end
