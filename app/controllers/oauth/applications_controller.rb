# frozen_string_literal: true

module Oauth
  class ApplicationsController < Oauth::ApplicationController
    before_action :authenticate_admin!

    def index
      @applications = current_user.oauth_applications.available
    end

    def create
      @application = Doorkeeper::Application.new(application_params)
      @application.owner = current_user

      if @application.save
        ga_client.events.create("oauth_applications", "create", ds: "web")
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
  end
end
