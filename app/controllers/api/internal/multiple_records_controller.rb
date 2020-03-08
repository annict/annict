# frozen_string_literal: true

module API
  module Internal
    class MultipleRecordsController < API::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        records = MultipleRecordsService.new(current_user)
        records.save!(params[:episode_ids])
        ga_client.page_category = params[:page_category]
        ga_client.events.create(:multiple_records, :create, ds: "internal_api")
        flash[:notice] = t "messages.multiple_records.create.saved"
        head 201
      end
    end
  end
end
