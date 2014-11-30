class ChannelWorksController < ApplicationController
  before_filter :authenticate_user!

  def index(page: nil)
    @works = current_user.works.wanna_watch_and_watching
              .program_registered
              .includes(:episodes)
              .order(released_at: :desc)
              .page(page)
  end
end
