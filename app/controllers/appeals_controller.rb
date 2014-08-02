class AppealsController < ApplicationController
  before_filter :authenticate_user!
  before_filter :set_work, only: [:create]


  def create
    AppealMailer.update_request(current_user, @work).deliver
    redirect_to :back, notice: t('appeals.sent_html')
  end
end