# frozen_string_literal: true

module Api
  module Internal
    class RecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i(destroy)

      def destroy
        record = current_user.records.only_kept.find(params[:record_id])

        RecordDestroyer.new(record: record).call

        head 204
      end
    end
  end
end
