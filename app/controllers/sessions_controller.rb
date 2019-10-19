# frozen_string_literal: true

class SessionsController < Devise::SessionsController
  def new
    store_location_for(:user, params[:back]) if params[:back].present?
    super
  end
end
