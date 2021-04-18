# frozen_string_literal: true

module My
  class SidebarController < My::ApplicationController
    layout false

    def show
      @user = current_user
    end
  end
end
