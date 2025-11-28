# typed: false
# frozen_string_literal: true

module Settings
  class TokensController < ApplicationV6Controller
    before_action :authenticate_user!

    def new
      @token = current_user.oauth_access_tokens.new
    end

    def create
      @token = current_user.oauth_access_tokens.new(access_token_params)

      if @token.save(context: :personal)
        flash[:notice] = t("messages.settings.tokens.created")
        flash[:created_token] = {id: @token.id, token: @token.token}
        redirect_to settings_app_list_path
      else
        render :new, status: :unprocessable_entity
      end
    end

    def edit
      @token = current_user.oauth_access_tokens.available.personal.find(params[:token_id])
    end

    def update
      @token = current_user.oauth_access_tokens.available.personal.find(params[:token_id])
      @token.attributes = access_token_params

      if @token.save
        flash[:notice] = t("messages.settings.tokens.updated")
        redirect_to settings_app_list_path
      else
        render :edit, status: :unprocessable_entity
      end
    end

    def destroy
      @token = current_user.oauth_access_tokens.available.personal.find(params[:token_id])

      @token.destroy

      flash[:notice] = t("messages.settings.tokens.deleted")
      redirect_to settings_app_list_path
    end

    private

    def access_token_params
      params.require(:oauth_access_token).permit(:description, :scopes)
    end
  end
end
