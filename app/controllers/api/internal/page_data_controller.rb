# frozen_string_literal: true

module Api
  module Internal
    class PageDataController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(index)

      def index(page_category, page_params = "{}")
        @page_category = page_category
        @page_params = JSON.parse(page_params)
        service_name = "PageData::#{@page_category.classify}Service"
        @results = service_name.safe_constantize.exec(@page_params)

        render page_category
      end
    end
  end
end
