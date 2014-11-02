class ShotsController < ApplicationController
  layout 'application_no_navbar'

  def show(username)
    @user = User.find_by(username: username)
  end
end
