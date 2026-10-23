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
      # `allowed_locales` が空のユーザーは、読み込み時にenumerizeが既定値を代入して
      # 未保存の変更を持ち、インスタンスのままではロックできない。そのため、
      # インスタンスを経由せずに行をロックし、行の存在と `deleted_at` を読む。
      saved = false
      owner_unavailable = false
      User.transaction do
        locked_owner = User.lock.where(id: owner.id).pick(:id, :deleted_at)
        owner_unavailable = locked_owner.nil? || locked_owner[1].present?
        saved = @application.save unless owner_unavailable
      end

      return head :forbidden if owner_unavailable

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
