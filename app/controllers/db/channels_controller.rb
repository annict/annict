# typed: false
# frozen_string_literal: true

module Db
  class ChannelsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @channels = Channel
        .without_deleted
        .eager_load(:channel_group)
        .merge(ChannelGroup.without_deleted)
        .order(vod: :desc, "channel_groups.sort_number": :asc, sort_number: :asc)
    end

    def new
      @channel = Channel.new
      authorize @channel
    end

    def create
      @channel = Channel.new(channel_params)
      authorize @channel

      return render(:new, status: :unprocessable_entity) unless @channel.valid?

      @channel.save

      redirect_to db_channel_list_path, notice: t("messages._common.created")
    end

    def edit
      @channel = Channel.without_deleted.find(params[:id])
      authorize @channel
    end

    def update
      @channel = Channel.without_deleted.find(params[:id])
      authorize @channel

      @channel.attributes = channel_params

      return render(:edit, status: :unprocessable_entity) unless @channel.valid?

      @channel.save

      redirect_to db_channel_list_path, notice: t("messages._common.updated")
    end

    def destroy
      @channel = Channel.without_deleted.find(params[:id])
      authorize @channel

      @channel.destroy_in_batches

      redirect_back(
        fallback_location: db_channel_list_path,
        notice: t("messages._common.deleted")
      )
    end

    private

    def channel_params
      params.require(:channel).permit(:name, :channel_group_id, :vod, :sort_number)
    end
  end
end
