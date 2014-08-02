class Api::WorksController < Api::ApplicationController
  before_action :authenticate_user!, only: [:hide]


  def hide(id)
    work = Work.find(id)
    current_user.hide(work)

    render status: 200, nothing: true
  end
end