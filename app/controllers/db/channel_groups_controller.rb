# typed: false
# frozen_string_literal: true

module Db
  class ChannelGroupsController < Db::ApplicationController
    before_action :authenticate_user!, only: %i[new create edit update destroy]

    def index
      @channel_groups = ChannelGroup
        .without_deleted
        .order(:sort_number)
    end

    def new
      @channel_group = ChannelGroup.new
      authorize @channel_group
    end

    def create
      @channel_group = ChannelGroup.new(channel_group_params)
      authorize @channel_group

      return render(:new, status: :unprocessable_entity) unless @channel_group.valid?

      @channel_group.save

      redirect_to db_channel_group_list_path, notice: t("messages._common.created")
    end

    def edit
      @channel_group = ChannelGroup.without_deleted.find(params[:id])
      authorize @channel_group
    end

    def update
      @channel_group = ChannelGroup.without_deleted.find(params[:id])
      authorize @channel_group

      @channel_group.attributes = channel_group_params

      return render(:edit, status: :unprocessable_entity) unless @channel_group.valid?

      @channel_group.save

      redirect_to db_channel_group_list_path, notice: t("messages._common.updated")
    end

    def destroy
      @channel_group = ChannelGroup.without_deleted.find(params[:id])
      authorize @channel_group

      @channel_group.destroy_in_batches

      redirect_back(
        fallback_location: db_channel_group_list_path,
        notice: t("messages._common.deleted")
      )
    end

    private

    def channel_group_params
      params.require(:channel_group).permit(:name, :sort_number)
    end
  end
end
