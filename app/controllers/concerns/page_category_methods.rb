# frozen_string_literal: true

module PageCategoryMethods
  extend ActiveSupport::Concern

  included do
    helper_method :page_category
  end

  private

  def page_category
    @page_category.presence || "other"
  end

  def store_page_category
    action = case params[:action]
    when "create" then "new"
    when "update" then "edit"
    else
      params[:action]
    end

    @page_category = "#{params[:controller]}_#{action}".tr("/", "_")

    page = gon.page.presence || {}
    page[:category] = page_category
    gon.push(page: page)
  end
end
