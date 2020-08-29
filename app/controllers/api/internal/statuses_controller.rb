# frozen_string_literal: true

module Api
  module Internal
    class StatusesController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      def select
        return head(:unauthorized) unless user_signed_in?

        @work = Work.only_kept.find(params[:work_id])

        UpdateStatusRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(anime: @work, kind: params[:status_kind])

        head 200
      end
    end
  end
end
