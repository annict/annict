# frozen_string_literal: true

class ChannelWorksController < ApplicationController
  before_action :authenticate_user!

  def index
    @works = current_user.
      works.
      wanna_watch_and_watching.
      without_deleted.
      slot_registered.
      includes(:episodes, :work_image).
      order_by_season(:desc)
  end
end
