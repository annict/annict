# frozen_string_literal: true
# == Schema Information
#
# Table name: providers
#
#  id               :integer          not null, primary key
#  user_id          :integer          not null
#  name             :string(510)      not null
#  uid              :string(510)      not null
#  token            :string(510)      not null
#  token_expires_at :integer
#  token_secret     :string(510)
#  created_at       :datetime
#  updated_at       :datetime
#
# Indexes
#
#  providers_name_uid_key  (name,uid) UNIQUE
#  providers_user_id_idx   (user_id)
#

class ProvidersController < ApplicationController
  before_action :authenticate_user!

  def destroy
    provider = current_user.providers.find(params[:id])

    ActiveRecord::Base.transaction do
      case provider.name
      when "twitter"
        provider.user.setting.update_column(:share_record_to_twitter, false)
      end

      provider.destroy
    end

    flash[:notice] = t("messages.providers.removed")
    redirect_back fallback_location: providers_path
  end
end
