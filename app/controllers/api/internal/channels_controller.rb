# frozen_string_literal: true

module Api
  module Internal
    class ChannelsController < Api::Internal::ApplicationController
      before_action :authenticate_user!
      before_action :set_work

      def select(channel_id)
        channel_work = current_user.channel_works.where(work: @work).first_or_initialize

        if channel_id == "no_select"
          channel_work.destroy if channel_work.present?
          return render(status: 200, nothing: true)
        end

        channel = Channel.find(channel_id)
        channel_work.channel = channel

        render status: 200, nothing: true if channel_work.save
      end
    end
  end
end
