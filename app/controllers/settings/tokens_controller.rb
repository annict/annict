# frozen_string_literal: true

module Settings
  class TokensController < ApplicationController
    permits :description, :scopes, model_name: "Doorkeeper::AccessToken"

    before_action :authenticate_user!

    def new
      @token = current_user.oauth_access_tokens.new
    end

    def create(doorkeeper_access_token)
      @token = current_user.oauth_access_tokens.new(doorkeeper_access_token)

      if @token.save(context: :personal)
        flash[:notice] = t("messages.settings.tokens.created")
        flash[:created_token] = { id: @token.id, token: @token.token }
        redirect_to settings_apps_path
      else
        render :new
      end
    end

    def edit(id)
      @token = current_user.oauth_access_tokens.available.personal.find(id)
    end

    def update(id, doorkeeper_access_token)
      @token = current_user.oauth_access_tokens.available.personal.find(id)
      @token.attributes = doorkeeper_access_token

      if @token.save(context: :personal)
        flash[:notice] = t("messages.settings.tokens.updated")
        redirect_to settings_apps_path
      else
        render :edit
      end
    end

    def destroy(id)
      @token = current_user.oauth_access_tokens.available.personal.find(id)

      @token.destroy

      flash[:notice] = t("messages.settings.tokens.deleted")
      redirect_to settings_apps_path
    end
  end
end
