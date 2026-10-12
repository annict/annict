# typed: false
# frozen_string_literal: true

module Oauth
  class ApplicationsController < Oauth::ApplicationController
    before_action :authenticate_admin!

    def index
      @applications = current_user.oauth_applications.available
    end

    def new
      @application = Oauth::Application.new
    end

    def create
      @application = Oauth::Application.new(application_params)
      owner = current_user
      @application.owner = owner

      # Go版の退会受付もユーザー行をロックする。ロック後に削除状態を読み直してから
      # 保存することで、退会とアプリケーション作成の順序を確定させる。
      saved = false
      owner.with_lock do
        saved = @application.save unless owner.deleted_at?
      end

      return head :forbidden if owner.deleted_at?

      if saved
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
        flash[:notice] = t "messages._common.updated"
        redirect_to oauth_application_url(@application)
      else
        render :edit
      end
    end

    def destroy
      @application = current_user.oauth_applications.available.find(params[:id])
      @application.destroy_in_batches
      flash[:notice] = t "messages._common.deleted"
      redirect_to oauth_applications_url
    end

    private

    def application_params
      if params.respond_to?(:permit)
        params.require(:oauth_application).permit(:name, :redirect_uri, :scopes)
      else
        params[:oauth_application]&.slice(:name, :redirect_uri, :scopes)
      end
    end
  end
end
