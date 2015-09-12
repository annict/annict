class Db::ApplicationController < ActionController::Base
  include Pundit
  include FlashMessage

  layout "db"

  rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

  before_action :set_ransack_params
  before_action :set_opened_edit_requests

  private

  def set_ransack_params
    @q = Work.search(params[:q])
  end

  def set_opened_edit_requests
    @opened_edit_requests = EditRequest.opened
  end

  def user_not_authorized
    flash[:alert] = "アクセスが許可されていません"
    redirect_to(request.referrer || db_root_path)
  end
end
