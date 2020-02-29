# frozen_string_literal: true

module Web
  class SessionsController < Devise::SessionsController
    def new
      store_location_for(:user, params[:back]) if params[:back]
      super
    end
  end
end
