# frozen_string_literal: true

module V4
  module PageCategorizable
    extend ActiveSupport::Concern

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
end
