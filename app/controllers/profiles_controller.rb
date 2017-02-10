# frozen_string_literal: true
# == Schema Information
#
# Table name: profiles
#
#  id                                  :integer          not null, primary key
#  user_id                             :integer          not null
#  name                                :string(510)      default(""), not null
#  description                         :string(510)      default(""), not null
#  created_at                          :datetime
#  updated_at                          :datetime
#  background_image_animated           :boolean          default(FALSE), not null
#  tombo_avatar_file_name              :string
#  tombo_avatar_content_type           :string
#  tombo_avatar_file_size              :integer
#  tombo_avatar_updated_at             :datetime
#  tombo_background_image_file_name    :string
#  tombo_background_image_content_type :string
#  tombo_background_image_file_size    :integer
#  tombo_background_image_updated_at   :datetime
#  url                                 :string
#
# Indexes
#
#  profiles_user_id_idx  (user_id)
#  profiles_user_id_key  (user_id) UNIQUE
#

class ProfilesController < ApplicationController
  permits :tombo_avatar, :tombo_background_image, :description, :name, :url

  before_action :authenticate_user!

  def update(profile)
    if current_user.profile.update_attributes(profile)
      redirect_to profile_path, notice: t("messages.profiles.saved")
    else
      render :show
    end
  end
end
