# frozen_string_literal: true

module Oauth
  class ApplicationsController < Oauth::ApplicationController
    before_action :authenticate_admin!

    def index
      @applications = current_user.oauth_applications.available
    end

    def new
      @application = Doorkeeper::Application.new
    end

    def create
      @application = Doorkeeper::Application.new(application_params)
      @application.owner = current_user

      if @application.save
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:oauth_applications, :create)
        flash[:notice] = t "messages.oauth.applications.created"
        redirect_to oauth_application_url(@application)
      else
        render :new
      end
    end

    def show
      @application = current_user.oauth_applications.available.find(params[:id])
    end

    def edit
      @application = current_user.oauth_applications.available.find(params[:id])
    end

    def update
      @application = current_user.oauth_applications.available.find(params[:id])
      if @application.update(application_params)
        flash[:notice] = t "doorkeeper.flash.applications.update.notice"
        redirect_to oauth_application_url(@application)
      else
        render :edit
      end
    end

    def destroy
      @application = current_user.oauth_applications.available.find(params[:id])
      @application.destroy_in_batches
      flash[:notice] = t "messages.oauth.applications.deleted"
      redirect_to oauth_applications_url
    end

    private

    def application_params
      if params.respond_to?(:permit)
        params.require(:doorkeeper_application).permit(:name, :redirect_uri, :scopes)
      else
        params[:doorkeeper_application]&.slice(:name, :redirect_uri, :scopes)
      end
    end
  end
end
