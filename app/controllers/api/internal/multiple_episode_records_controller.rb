# frozen_string_literal: true

module Api
  module Internal
    class MultipleEpisodeRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        job = BulkCreateEpisodeRecordsJob.perform_later(current_user.id, params[:episode_ids])

        render(status: 201, json: { job_id: job.provider_job_id })
      end
    end
  end
end
