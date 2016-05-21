# frozen_string_literal: true

class Oauth::ApplicationsController < Doorkeeper::ApplicationsController
  include ViewSelector

  layout "v3/application"

  before_action :set_search_params

  def index
    @applications = current_user.oauth_applications.available
  end

  def create
    @application = Doorkeeper::Application.new(application_params)
    @application.owner = current_user

    if @application.save
      flash[:notice] = "登録しました"
      redirect_to oauth_application_url(@application)
    else
      render :new
    end
  end

  def destroy
    @application.hide!
    redirect_to oauth_applications_url
  end

  private

  def set_application
    @application = current_user.oauth_applications.available.find(params[:id])
  end

  def set_search_params
    @search = SearchService.new(params[:q])
  end
end
