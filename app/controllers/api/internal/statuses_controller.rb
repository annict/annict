# frozen_string_literal: true

module Api
  module Internal
    class StatusesController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!

      def select
        @work = Work.only_kept.find(params[:work_id])

        UpdateStatusRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).create(work: @work, kind: params[:status_kind])

        head 200
      end
    end
  end
end
