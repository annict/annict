# frozen_string_literal: true

module Api
  module V1
    module Me
      class StatusesController < Api::V1::ApplicationController
        include V4::GraphqlRunnable

        before_action :prepare_params!, only: [:create]

        def create
          work = Work.only_kept.find(@params.work_id)

          UpdateStatusRepository.new(
            graphql_client: graphql_client(viewer: current_user)
          ).create(anime: work, kind: @params.kind)

          head 204
        end
      end
    end
  end
end
