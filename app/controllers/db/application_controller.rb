class Db::ApplicationController < ActionController::Base
  layout "db"

  before_action :authenticate_user!
  before_action :set_ransack_params

  private

  def set_ransack_params
    @q = Work.search(params[:q])
  end
end
