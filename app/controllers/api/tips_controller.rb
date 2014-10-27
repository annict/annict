class Api::TipsController < ApplicationController
  before_filter :authenticate_user!, only: [:finish]

  def finish(partial_name)
    tip = Tip.find_by(partial_name: partial_name)
    current_user.finish_tip!(tip)

    render status: 200, nothing: true
  end
end
