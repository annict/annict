# frozen_string_literal: true

module Api
  module Internal
    class MultipleRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        CreateMultipleEpisodeRecordsJob.perform_later(current_user.id, params[:episode_ids])

        flash[:notice] = t "messages.multiple_records.create.saved"
        head 201
      end
    end
  end
end
