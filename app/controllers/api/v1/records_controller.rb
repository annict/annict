# frozen_string_literal: true

module Api
  module V1
    class RecordsController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @records = Record.published.includes(episode: :work, user: :profile).all
        @records = Api::V1::RecordIndexService.new(@records, @params).result
      end
    end
  end
end
