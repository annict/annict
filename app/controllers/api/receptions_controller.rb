class Api::ReceptionsController < ApplicationController
  before_filter :authenticate_user!
  before_filter :set_channel, only: [:create, :destroy]


  def create
    current_user.receive(@channel)

    render status: 200, nothing: true
  end

  def destroy
    current_user.unreceive(@channel)

    render status: 200, nothing: true
  end


  private

  def set_channel
    channel_id = params[:id] || params[:channel_id]
    @channel = Channel.find(channel_id)
  end
end