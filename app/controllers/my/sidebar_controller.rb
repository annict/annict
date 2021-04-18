# frozen_string_literal: true

module My
  class SidebarController < My::ApplicationController
    layout false

    before_action :authenticate_user!, only: %i(show)

    def show
      @user = current_user
    end
  end
end
