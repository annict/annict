class Marie::ApplicationController < ActionController::Base
  layout 'marie'

  before_action :authenticate_user!

  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end
