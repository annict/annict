# frozen_string_literal: true

module DB
  class ChannelsController < DB::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update hide destroy)

    def index
      @channels = Channel.
        all.
        preload(:channel_group).
        order(vod: :desc, channel_group_id: :asc, id: :desc)
    end

    def new
      @channel = Channel.new
      authorize @channel, :new?
    end

    def create
      @channel = Channel.new(channel_params)
      authorize @channel, :create?

      if @channel.save
        redirect_to db_channels_path, notice: t("messages._common.created")
      else
        render :new
      end
    end

    def edit
      @channel = Channel.find(params[:id])
      authorize @channel, :edit?
    end

    def update
      @channel = Channel.find(params[:id])
      authorize @channel, :update?

      @channel.attributes = channel_params

      if @channel.save
        redirect_to db_channels_path, notice: t("messages._common.updated")
      else
        render :edit
      end
    end

    def hide
      @channel = Channel.find(params[:id])
      authorize @channel, :hide?

      @channel.soft_delete_with_children

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_channels_path
    end

    def destroy
      @channel = Channel.find(params[:id])
      authorize @channel, :destroy?

      @channel.destroy

      flash[:notice] = t("messages._common.deleted")
      redirect_back fallback_location: db_channels_path
    end

    private

    def channel_params
      params.require(:channel).permit(:name, :channel_group_id, :vod)
    end
  end
end
