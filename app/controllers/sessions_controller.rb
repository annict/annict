# frozen_string_literal: true
# == Schema Information
#
# Table name: sessions
#
#  id         :integer          not null, primary key
#  session_id :string(510)      not null
#  data       :text
#  created_at :datetime
#  updated_at :datetime
#
# Indexes
#
#  sessions_session_id_key  (session_id) UNIQUE
#

class SessionsController < Devise::SessionsController
  def new
    store_location_for(:user, params[:back]) if params[:back].present?
    super
  end
end
