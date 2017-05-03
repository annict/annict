# frozen_string_literal: true

module Db
  class ApplicationController < ActionController::Base
    include Pundit

    include ControllerCommon
    include FlashMessage
    include Analyzable
    include Gonable

    layout "db"

    rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

    helper_method :client_uuid, :gon

    before_action :redirect_if_unexpected_subdomain
    before_action :switch_languages
    before_action :set_search_params
    before_action :load_data_into_gon

    private

    def load_character
      @character = Character.find(params[:character_id])
    end

    def load_work
      @work = Work.find(params[:work_id])
    end

    def load_series
      @series = Series.find(params[:series_id])
    end

    def set_search_params
      @search = SearchService.new(params[:q], scope: :all)
    end

    def user_not_authorized
      flash[:alert] = t "messages._common.you_can_not_access_there"
      redirect_to(request.referrer || db_root_path)
    end
  end
end
