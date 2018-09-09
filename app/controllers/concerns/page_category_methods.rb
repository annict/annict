# frozen_string_literal: true

module PageCategoryMethods
  extend ActiveSupport::Concern

  included do
    helper_method :page_category
  end

  private

  def page_category
    @page_category ||= begin
      action = case params[:action]
      when "create" then "new"
      when "update" then "edit"
      else
        params[:action]
      end

      "#{params[:controller]}_#{action}".tr("/", "_")
    end
  end
end
