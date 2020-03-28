# frozen_string_literal: true

module V4
  module Db
    class ApplicationController < ActionController::Base
      include Pundit

      include V4::RavenContext
      include V4::Loggable
      include V4::Localizable
      include V4::PageCategorizable

      layout "v4"

      helper_method :client_uuid, :page_category

      before_action :set_raven_context
      before_action :set_search_params
      around_action :set_locale

      private

      def set_search_params
        @search = SearchService.new(params[:q], scope: :all)
      end
    end
  end
end
