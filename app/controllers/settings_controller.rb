# frozen_string_literal: true
# == Schema Information
#
# Table name: settings
#
#  id                       :integer          not null, primary key
#  user_id                  :integer          not null
#  hide_record_comment     :boolean          default(TRUE), not null
#  created_at               :datetime         not null
#  updated_at               :datetime         not null
#  share_record_to_twitter  :boolean          default(FALSE)
#  share_record_to_facebook :boolean          default(FALSE)
#
# Indexes
#
#  index_settings_on_user_id  (user_id)
#

class SettingsController < ApplicationController
  before_action :authenticate_user!

  def index
    return render(:index) unless device_pc?
    redirect_to profile_path
  end
end
