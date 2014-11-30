class Api::UserProgramsController < ApplicationController
  before_filter :authenticate_user!


  def index(page: nil)
    @programs = current_user.programs.unchecked
                  .where('started_at < ?', Date.tomorrow + 1.day + 5.hours)
                  .includes(:channel, :work, episode: [:work])
                  .order(started_at: :desc)
                  .page(page)
  end
end
