# frozen_string_literal: true

module DB
  class ChannelGroupsController < DB::ApplicationController
    before_action :authenticate_user!, only: %i(new create edit update unpublish destroy)

    def index
      @channel_groups = ChannelGroup.
        all.
        order(sort_number: :desc)
    end

    def new
      @channel_group = ChannelGroup.new
      authorize(@channel_group, :new?)
    end

    def create
      @channel_group = ChannelGroup.new(channel_group_params)
      authorize(@channel_group, :create?)

      if @channel_group.save
        redirect_to db_channel_groups_path, notice: t("messages._common.created")
      else
        render :new
      end
    end

    def edit
      @channel_group = ChannelGroup.find(params[:id])
      authorize(@channel_group, :edit?)
    end

    def update
      @channel_group = ChannelGroup.find(params[:id])
      authorize(@channel_group, :update?)

      @channel_group.attributes = channel_group_params

      if @channel_group.save
        redirect_to db_channel_groups_path, notice: t("messages._common.updated")
      else
        render :edit
      end
    end

    def unpublish
      @channel_group = ChannelGroup.find(params[:id])
      authorize(@channel_group, :unpublish?)

      @channel_group.soft_delete

      flash[:notice] = t("messages._common.unpublished")
      redirect_back fallback_location: db_channel_groups_path
    end

    def destroy
      @channel_group = ChannelGroup.find(params[:id])
      authorize(@channel_group, :destroy?)

      @channel_group.destroy

      flash[:notice] = t("messages._common.deleted")
      redirect_back fallback_location: db_channel_groups_path
    end

    private

    def channel_group_params
      params.require(:channel_group).permit(:name, :sort_number)
    end
  end
end
