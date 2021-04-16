# frozen_string_literal: true

class SidebarProfilesController < V4::ApplicationController
  layout false

  def show
    @user = current_user
  end
end
