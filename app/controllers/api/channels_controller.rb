class Api::ChannelsController < Api::ApplicationController
  before_filter :authenticate_user!
  before_filter :set_work


  def select(channel_id)
    channel = Channel.find(channel_id)
    channel_work = current_user.channel_works.where(work: @work).first_or_initialize
    channel_work.channel = channel

    if channel_work.save
      render status: 200, nothing: true
    end
  end
end