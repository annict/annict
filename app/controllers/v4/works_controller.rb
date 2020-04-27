# frozen_string_literal: true

module V4
  class WorksController < V4::ApplicationController
    def show
      @work = WorkDetail::WorkRepository.new(graphql_client: graphql_client).fetch(work_id: params[:id])
      @vod_channels = WorkDetail::VodChannelsRepository.new(graphql_client: graphql_client).fetch(work: @work)
      @existing_vod_channels = @vod_channels.select { |vod_channel| vod_channel.programs.first.present? }
    end
  end
end
