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

  def destroy(id)
    provider = current_user.providers.find(id)

    ActiveRecord::Base.transaction do
      case provider.name
      when "twitter"
        provider.user.setting.update_column(:share_record_to_twitter, false)
      when "facebook"
        provider.user.setting.update_column(:share_record_to_facebook, false)
      end

      provider.destroy
    end

    redirect_to :back, notice: "連携を解除しました"
  end
end
