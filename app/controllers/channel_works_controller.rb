class ChannelWorksController < ApplicationController
  before_filter :authenticate_user!

  def index(page: nil)
    @works = current_user.wanna_watch_or_watching_works
              .program_registered
              .includes(:episodes)
              .order(released_at: :desc)
              .page(page)
  end
end
