class Api::UserProgramsController < ApplicationController
  before_filter :authenticate_user!


  def index(page: nil)
    @programs = current_user.unchecked_programs
                  .where('started_at < ?', Date.tomorrow + 1.day + 5.hours)
                  .includes(:channel, episode: [:work])
                  .order(started_at: :desc)
                  .page(page)
  end
end
