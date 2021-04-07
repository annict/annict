# frozen_string_literal: true

module My
  class SidebarProfilesController < My::ApplicationController
    def show
      @user = current_user
    end
  end
end
