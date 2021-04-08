# frozen_string_literal: true

module Frame
  class SidebarProfilesController < Frame::ApplicationController
    def show
      @user = current_user
    end
  end
end
