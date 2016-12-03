# frozen_string_literal: true

module Db
  class ApplicationController < ActionController::Base
    include Pundit
    include ControllerCommon
    include FlashMessage

    layout "db"

    rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

    helper_method :client_uuid, :gon
    before_action :set_search_params

    private

    def set_search_params
      @search = SearchService.new(params[:q], scope: :all)
    end

    def user_not_authorized
      flash[:alert] = "アクセスが許可されていません"
      redirect_to(request.referrer || db_root_path)
    end
  end
end
