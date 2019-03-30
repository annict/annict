# frozen_string_literal: true

module Api
  module Internal
    class PageDataController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(index)

      def index
        @page_category = params[:page_category]
        @page_params = JSON.parse(params[:page_params].presence || "{}")
        service_name = "PageData::#{@page_category.classify}Service"
        @results = service_name.safe_constantize.exec(current_user, @page_params)

        render page_category
      end
    end
  end
end
