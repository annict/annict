# frozen_string_literal: true

module Db
  class ApplicationController < ActionController::Base
    include Pundit
    include FlashMessage

    layout "db"

    rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

    before_action :set_opened_edit_requests

    private

    def set_search_params
      @search = SearchService.new(params[:q], scope: :all)
    end

    def set_opened_edit_requests
      @opened_edit_requests = EditRequest.opened
    end

    def user_not_authorized
      flash[:alert] = "アクセスが許可されていません"
      redirect_to(request.referrer || db_root_path)
    end
  end
end
