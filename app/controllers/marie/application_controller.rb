class Marie::ApplicationController < ActionController::Base
  layout 'marie'

  before_filter :authenticate_staff!


  private

  def set_work
    @work = Work.find(params[:work_id])
  end
end