# frozen_string_literal: true

module Api
  module Internal
    class ChannelsController < Api::Internal::ApplicationController
      before_action :authenticate_user!
      before_action :load_work

      def select
        channel_work = current_user.channel_works.where(work: @work).first_or_initialize

        if params[:channel_id] == "no_select"
          channel_work.destroy if channel_work.present?
          return head(200)
        end

        channel = Channel.published.find(params[:channel_id])
        channel_work.channel = channel

        head(200) if channel_work.save
      end
    end
  end
end
