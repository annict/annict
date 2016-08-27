class MutesController < ApplicationController
  before_action :authenticate_user!

  def index
    @mute_users = current_user.mute_users.order(id: :desc)
    render layout: "v1/application"
  end
end
