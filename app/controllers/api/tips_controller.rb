class Api::TipsController < ApplicationController
  before_filter :authenticate_user!, only: [:finish]

  def finish(partial_name)
    UserTipsService.new(current_user).finish!(partial_name)

    render status: 200, nothing: true
  end
end
